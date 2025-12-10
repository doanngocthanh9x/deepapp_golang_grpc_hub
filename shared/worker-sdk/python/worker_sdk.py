#!/usr/bin/env python3
"""
Worker SDK - Base class for creating workers with plugin support
Provides easy worker-to-worker communication and capability registration
"""

import grpc
import json
import time
import uuid
import threading
import queue
from datetime import datetime
from typing import Dict, Callable, Any, Optional, List
from abc import ABC, abstractmethod


class WorkerSDK(ABC):
    """
    Base SDK for creating gRPC workers with built-in worker-to-worker communication
    
    Usage:
        class MyWorker(WorkerSDK):
            def register_capabilities(self):
                self.add_capability(
                    name="my_task",
                    handler=self.handle_my_task,
                    description="Does something useful",
                    http_method="POST"
                )
            
            def handle_my_task(self, params: dict) -> dict:
                # Call another worker if needed
                result = self.call_worker("other-worker", "other_task", {"key": "value"})
                return {"status": "success", "result": result}
        
        worker = MyWorker("my-worker", "localhost:50051")
        worker.run()
    """
    
    def __init__(self, worker_id: str, hub_address: str, worker_type: str = "python"):
        self.worker_id = worker_id
        self.hub_address = hub_address
        self.worker_type = worker_type
        self.channel = None
        self.stub = None
        self.running = False
        self.send_queue = queue.Queue()
        
        # Capability registry
        self.capabilities = {}
        self.capability_handlers: Dict[str, Callable] = {}
        
        # Worker-to-worker call tracking
        self.pending_calls = {}
        self.pending_lock = threading.Lock()
        
        # Import proto files (must be in same directory or PYTHONPATH)
        try:
            import hub_pb2
            import hub_pb2_grpc
            self.hub_pb2 = hub_pb2
            self.hub_pb2_grpc = hub_pb2_grpc
        except ImportError as e:
            raise RuntimeError(
                "Proto files not found. Please generate them with:\n"
                "python3 -m grpc_tools.protoc -I./proto --python_out=. --grpc_python_out=. ./proto/hub.proto"
            ) from e
    
    @abstractmethod
    def register_capabilities(self):
        """
        Override this method to register your worker's capabilities
        Use self.add_capability() to register each capability
        """
        pass
    
    def add_capability(
        self,
        name: str,
        handler: Callable[[dict], dict],
        description: str = "",
        http_method: str = "POST",
        accepts_file: bool = False,
        file_field_name: str = "",
        input_schema: str = "{}",
        output_schema: str = "{}"
    ):
        """
        Register a capability handler
        
        Args:
            name: Capability name (e.g., "process_data")
            handler: Function that takes dict params and returns dict result
            description: Human-readable description
            http_method: HTTP method for REST API (GET, POST, etc.)
            accepts_file: Whether this capability accepts file uploads
            file_field_name: Name of the file field if accepts_file=True
            input_schema: JSON schema for input validation
            output_schema: JSON schema for output
        """
        self.capability_handlers[name] = handler
        self.capabilities[name] = {
            "name": name,
            "description": description,
            "input_schema": input_schema,
            "output_schema": output_schema,
            "http_method": http_method,
            "accepts_file": accepts_file,
        }
        if file_field_name:
            self.capabilities[name]["file_field_name"] = file_field_name
        
        self.log(f"✓ Registered capability: {name}")
    
    def call_worker(
        self,
        target_worker: str,
        capability: str,
        params: dict,
        timeout: int = 30
    ) -> dict:
        """
        Call another worker's capability through the Hub
        
        Args:
            target_worker: Worker ID to call (e.g., "java-worker")
            capability: Capability name on the target worker
            params: Parameters to send (will be JSON serialized)
            timeout: Timeout in seconds
        
        Returns:
            dict: Response from the target worker
        
        Raises:
            TimeoutError: If no response within timeout
            Exception: If error response received
        """
        if not self.running:
            raise RuntimeError("Worker not connected. Call run() first")
        
        request_id = str(uuid.uuid4())
        
        self.log(f"🔗 Calling {target_worker}.{capability}")
        
        # Create worker call message
        call_msg = self.hub_pb2.Message(
            id=request_id,
            to=target_worker,
            channel=capability,
            content=json.dumps(params),
            timestamp=datetime.now().isoformat(),
            type=self.hub_pb2.WORKER_CALL
        )
        setattr(call_msg, 'from', self.worker_id)
        call_msg.metadata['capability'] = capability
        
        # Register pending call
        response_event = threading.Event()
        response_data = {'response': None, 'error': None}
        
        with self.pending_lock:
            self.pending_calls[request_id] = {
                'event': response_event,
                'data': response_data
            }
        
        # Send the call
        self.send_queue.put(call_msg)
        
        # Wait for response
        if response_event.wait(timeout=timeout):
            if response_data['error']:
                raise Exception(response_data['error'])
            return response_data['response']
        else:
            # Timeout - clean up
            with self.pending_lock:
                self.pending_calls.pop(request_id, None)
            raise TimeoutError(f"No response from {target_worker} after {timeout}s")
    
    def _handle_worker_call_response(self, msg):
        """Internal: Handle response from worker-to-worker call"""
        request_id = msg.metadata.get('request_id', '')
        
        with self.pending_lock:
            if request_id and request_id in self.pending_calls:
                call_info = self.pending_calls[request_id]
                try:
                    response_content = json.loads(msg.content)
                    call_info['data']['response'] = response_content
                    call_info['event'].set()
                    # Clean up
                    del self.pending_calls[request_id]
                except Exception as e:
                    call_info['data']['error'] = f"Failed to parse response: {e}"
                    call_info['event'].set()
                    del self.pending_calls[request_id]
    
    def _process_message(self, msg):
        """Internal: Process incoming message"""
        channel = msg.channel
        
        if channel not in self.capability_handlers:
            return json.dumps({
                "error": f"Unknown capability: {channel}",
                "status": "failed"
            })
        
        try:
            # Parse input
            params = json.loads(msg.content) if msg.content else {}
            
            # Call handler
            result = self.capability_handlers[channel](params)
            
            # Ensure result is dict
            if not isinstance(result, dict):
                result = {"result": result}
            
            return json.dumps(result)
            
        except Exception as e:
            self.log(f"✗ Error in {channel}: {e}")
            return json.dumps({
                "error": str(e),
                "status": "failed"
            })
    
    def _send_registration(self):
        """Internal: Send registration message to Hub"""
        capabilities_list = list(self.capabilities.values())
        
        registration_data = {
            "worker_id": self.worker_id,
            "worker_type": self.worker_type,
            "capabilities": capabilities_list,
            "metadata": {
                "version": "1.0.0",
                "sdk_version": "2.0.0"
            }
        }
        
        register_msg = self.hub_pb2.Message(
            id=f"register-{int(time.time() * 1000000)}",
            to="hub",
            channel="system",
            content=json.dumps(registration_data),
            timestamp=datetime.now().isoformat(),
            type=self.hub_pb2.REGISTER
        )
        setattr(register_msg, 'from', self.worker_id)
        
        return register_msg
    
    def _request_generator(self):
        """Internal: Generate requests to send to Hub"""
        try:
            # Send registration first
            yield self._send_registration()
            self.log("📤 Sent registration")
            
            # Then send messages from queue
            while self.running:
                try:
                    msg = self.send_queue.get(block=True, timeout=1.0)
                    yield msg
                except queue.Empty:
                    continue
                except Exception as e:
                    self.log(f"✗ Generator error: {e}")
                    break
        except Exception as e:
            self.log(f"✗ Request generator error: {e}")
        finally:
            self.log("Generator exiting")
    
    def _receive_loop(self, response_stream):
        """Internal: Receive and process messages from Hub"""
        try:
            for msg in response_stream:
                if not self.running:
                    break
                
                try:
                    msg_from = getattr(msg, 'from')
                    msg_type = msg.type
                    
                    # Handle different message types
                    if msg_type == self.hub_pb2.RESPONSE:
                        # Response from worker-to-worker call
                        self._handle_worker_call_response(msg)
                        continue
                    
                    elif msg_type == self.hub_pb2.WORKER_CALL:
                        # Another worker calling us
                        response_content = self._process_message(msg)
                        
                        response_msg = self.hub_pb2.Message(
                            id=f"resp-{int(time.time() * 1000000)}",
                            to=msg_from,
                            channel=msg.channel,
                            content=response_content,
                            timestamp=datetime.now().isoformat(),
                            type=self.hub_pb2.RESPONSE
                        )
                        setattr(response_msg, 'from', self.worker_id)
                        response_msg.metadata['request_id'] = msg.id
                        response_msg.metadata['status'] = 'success'
                        self.send_queue.put(response_msg)
                    
                    else:
                        # Regular REQUEST
                        response_content = self._process_message(msg)
                        
                        response_msg = self.hub_pb2.Message(
                            id=f"resp-{int(time.time() * 1000000)}",
                            to=msg_from,
                            channel=msg.channel,
                            content=response_content,
                            timestamp=datetime.now().isoformat(),
                            type=self.hub_pb2.RESPONSE
                        )
                        setattr(response_msg, 'from', self.worker_id)
                        self.send_queue.put(response_msg)
                
                except Exception as e:
                    self.log(f"✗ Error processing message: {e}")
        
        except grpc.RpcError as e:
            if e.code() != grpc.StatusCode.CANCELLED:
                self.log(f"✗ gRPC Error: {e.code()} - {e.details()}")
        except Exception as e:
            self.log(f"✗ Receive loop error: {e}")
        finally:
            self.running = False
            self.log("Receive loop exited")
    
    def run(self):
        """Start the worker and connect to Hub"""
        self.log(f"🚀 Starting Worker")
        self.log(f"   ID: {self.worker_id}")
        self.log(f"   Hub: {self.hub_address}")
        self.log(f"=" * 50)
        
        # Register capabilities
        self.register_capabilities()
        self.log(f"✓ Registered {len(self.capabilities)} capabilities")
        
        try:
            # Connect to Hub
            self.log("Connecting to Hub...")
            self.channel = grpc.insecure_channel(self.hub_address)
            grpc.channel_ready_future(self.channel).result(timeout=10)
            self.log("✓ Connected to Hub")
            
            # Create stub
            self.stub = self.hub_pb2_grpc.HubServiceStub(self.channel)
            
            # Start bidirectional streaming
            self.running = True
            response_stream = self.stub.Connect(self._request_generator())
            
            self.log("✓ Registered with Hub")
            self.log("📨 Listening for requests...\n")
            
            # Start receive thread
            receive_thread = threading.Thread(target=self._receive_loop, args=(response_stream,), daemon=True)
            receive_thread.start()
            
            # Keep main thread alive
            self.log("Worker running. Press Ctrl+C to stop.\n")
            while self.running:
                time.sleep(1)
        
        except grpc.RpcError as e:
            self.log(f"✗ gRPC Error: {e.code()} - {e.details()}")
        except Exception as e:
            self.log(f"✗ Error: {e}")
        finally:
            self.stop()
    
    def stop(self):
        """Stop the worker"""
        self.running = False
        if self.channel:
            self.channel.close()
            self.log("✗ Disconnected from Hub")
    
    def log(self, message: str):
        """Log a message"""
        print(f"[{self.worker_id}] {message}")
