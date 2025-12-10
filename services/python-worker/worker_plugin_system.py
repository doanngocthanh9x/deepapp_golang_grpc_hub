#!/usr/bin/env python3
"""
Python Worker with Plugin System
Auto-discovers and loads all plugins from plugins/ directory
"""

import grpc
import json
import time
import sys
import os
import uuid
import threading
import queue
import traceback
from datetime import datetime

# Import generated proto files
import hub_pb2
import hub_pb2_grpc

# Import plugin system
from plugin_manager import PluginManager


class PluginWorker:
    """Worker that uses plugin system for capability management"""
    
    def __init__(self, worker_id='python-worker', hub_address='localhost:50051'):
        self.worker_id = worker_id
        self.hub_address = hub_address
        self.channel = None
        self.stub = None
        self.running = False
        self.send_queue = None
        
        # Worker-to-worker call tracking
        self.pending_calls = {}
        self.pending_lock = threading.Lock()
        
        # Plugin system
        self.plugin_manager = PluginManager()
        self.plugins = {}
        
    def load_plugins(self):
        """Auto-discover and load all plugins"""
        self.plugins = self.plugin_manager.load_all_plugins()
        
        if not self.plugins:
            print("⚠️  No plugins loaded! Worker will have no capabilities.")
        
        return self.plugins
    
    def call_worker(self, target_worker, capability, params, timeout=30):
        """
        Call another worker's capability through Hub
        
        Args:
            target_worker: ID of the target worker
            capability: Capability name to call
            params: Dictionary of parameters
            timeout: Timeout in seconds
            
        Returns:
            Response data as dictionary
        """
        if not self.send_queue:
            raise RuntimeError("Worker not connected")
        
        request_id = str(uuid.uuid4())
        print(f"🔗 Calling worker '{target_worker}' capability '{capability}'")
        
        # Create WORKER_CALL message
        call_msg = hub_pb2.Message(
            id=request_id,
            to=target_worker,
            channel=capability,
            content=json.dumps(params),
            timestamp=datetime.now().isoformat(),
            type=hub_pb2.WORKER_CALL
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
            with self.pending_lock:
                self.pending_calls.pop(request_id, None)
            
            if response_data['error']:
                raise Exception(response_data['error'])
            
            return response_data['response']
        else:
            with self.pending_lock:
                self.pending_calls.pop(request_id, None)
            raise TimeoutError(f"No response from {target_worker} after {timeout}s")
    
    def _handle_worker_call_response(self, msg):
        """Handle response from worker-to-worker call"""
        request_id = msg.metadata.get('request_id', '')
        
        with self.pending_lock:
            if request_id and request_id in self.pending_calls:
                call_info = self.pending_calls[request_id]
                try:
                    response_content = json.loads(msg.content)
                    call_info['data']['response'] = response_content
                    call_info['event'].set()
                except Exception as e:
                    call_info['data']['error'] = f"Failed to parse response: {e}"
                    call_info['event'].set()
    
    def process_message(self, msg):
        """Process incoming message using plugin system"""
        capability_name = msg.channel
        
        print(f"  → Processing capability: {capability_name}")
        
        # Get plugin for this capability
        plugin = self.plugin_manager.get_plugin(capability_name)
        
        if plugin is None:
            error_data = {
                "error": f"Unknown capability: {capability_name}",
                "status": "failed"
            }
            return json.dumps(error_data)
        
        try:
            # Parse input parameters
            try:
                params = json.loads(msg.content) if msg.content else {}
            except:
                params = {}
            
            # Execute plugin (pass self as worker_sdk for worker-to-worker calls)
            result = plugin.execute(params, worker_sdk=self)
            
            return json.dumps(result)
            
        except Exception as e:
            print(f"  ✗ Plugin execution error: {e}")
            traceback.print_exc()
            
            error_data = {
                "error": str(e),
                "status": "failed",
                "capability": capability_name
            }
            return json.dumps(error_data)
    
    def connect_and_run(self):
        """Connect to Hub and start processing messages"""
        # Load plugins first
        self.load_plugins()
        
        try:
            print(f"Connecting to gRPC Hub...")
            self.channel = grpc.insecure_channel(self.hub_address)
            grpc.channel_ready_future(self.channel).result(timeout=10)
            print(f"✓ Connected to Hub")
            
            self.stub = hub_pb2_grpc.HubServiceStub(self.channel)
            
            # Create message queue
            send_queue = queue.Queue()
            self.send_queue = send_queue
            
            # Generator function for sending messages
            def request_generator():
                try:
                    # Send registration
                    capabilities = self.plugin_manager.get_all_capabilities()
                    
                    registration_data = {
                        "worker_id": self.worker_id,
                        "worker_type": "python",
                        "capabilities": capabilities,
                        "metadata": {
                            "version": "2.0.0",
                            "description": "Python worker with plugin system",
                            "plugin_count": len(self.plugins)
                        }
                    }
                    
                    register_msg = hub_pb2.Message(
                        id=f"register-{int(time.time() * 1000000)}",
                        to="hub",
                        channel="system",
                        content=json.dumps(registration_data),
                        timestamp=datetime.now().isoformat(),
                        type=hub_pb2.REGISTER
                    )
                    setattr(register_msg, 'from', self.worker_id)
                    yield register_msg
                    print(f"📤 Sent registration with {len(capabilities)} capabilities")
                    
                    # Keep generator alive
                    while self.running:
                        try:
                            msg = send_queue.get(timeout=1)
                            yield msg
                        except queue.Empty:
                            continue
                            
                except Exception as e:
                    print(f"Generator error: {e}")
                    traceback.print_exc()
            
            # Start bidirectional streaming
            print(f"📡 Starting bidirectional stream...")
            response_stream = self.stub.Connect(request_generator())
            
            print(f"✓ Registered with Hub as '{self.worker_id}'")
            print(f"📨 Listening for requests with {len(self.plugins)} plugins loaded...\n")
            
            self.running = True
            
            # Receive loop
            def receive_loop():
                try:
                    for msg in response_stream:
                        if not self.running:
                            break
                        
                        try:
                            msg_from = getattr(msg, 'from')
                            msg_type = msg.type
                            
                            print(f"📬 Received message:")
                            print(f"   From: {msg_from}, Type: {msg_type}, Channel: {msg.channel}")
                            
                            # Handle different message types
                            if msg_type == hub_pb2.RESPONSE:
                                self._handle_worker_call_response(msg)
                                continue
                            
                            elif msg_type == hub_pb2.WORKER_CALL:
                                # Another worker calling us
                                print(f"   → Worker-to-worker call")
                                response_content = self.process_message(msg)
                                
                                response_msg = hub_pb2.Message(
                                    id=f"resp-{int(time.time() * 1000000)}",
                                    to=msg_from,
                                    channel=msg.channel,
                                    content=response_content,
                                    timestamp=datetime.now().isoformat(),
                                    type=hub_pb2.RESPONSE
                                )
                                setattr(response_msg, 'from', self.worker_id)
                                response_msg.metadata['request_id'] = msg.id
                                send_queue.put(response_msg)
                                print(f"   ✓ Queued response\n")
                            
                            else:
                                # Regular REQUEST
                                print(f"   → Regular request")
                                response_content = self.process_message(msg)
                                
                                response_msg = hub_pb2.Message(
                                    id=f"resp-{int(time.time() * 1000000)}",
                                    to=msg_from,
                                    channel=msg.channel,
                                    content=response_content,
                                    timestamp=datetime.now().isoformat(),
                                    type=hub_pb2.RESPONSE
                                )
                                setattr(response_msg, 'from', self.worker_id)
                                send_queue.put(response_msg)
                                print(f"   ✓ Queued response\n")
                        
                        except Exception as e:
                            print(f"✗ Error processing message: {e}")
                            traceback.print_exc()
                
                except grpc.RpcError as e:
                    if e.code() != grpc.StatusCode.CANCELLED:
                        print(f"✗ gRPC Error: {e.code()} - {e.details()}")
                except Exception as e:
                    print(f"✗ Receive loop error: {e}")
                    traceback.print_exc()
                finally:
                    self.running = False
            
            # Start receive thread
            receive_thread = threading.Thread(target=receive_loop, daemon=True)
            receive_thread.start()
            
            # Keep main thread alive
            print("Worker running. Press Ctrl+C to stop.\n")
            while self.running:
                time.sleep(1)
        
        except grpc.RpcError as e:
            print(f"\n✗ gRPC Error: {e.code()} - {e.details()}")
        except Exception as e:
            print(f"\n✗ Connection error: {e}")
            traceback.print_exc()
        finally:
            self.running = False
            if self.channel:
                self.channel.close()
            self.plugin_manager.unload_all_plugins()


def main():
    print("🐍 Python Worker (Plugin System)")
    print("=" * 50)
    
    worker_id = os.getenv('WORKER_ID', 'python-worker')
    hub_address = os.getenv('HUB_ADDRESS', 'localhost:50051')
    
    print(f"Worker ID: {worker_id}")
    print(f"Hub Address: {hub_address}")
    print("=" * 50)
    
    worker = PluginWorker(worker_id, hub_address)
    
    try:
        worker.connect_and_run()
    except KeyboardInterrupt:
        print("\n\n✗ Shutting down...")
        worker.running = False


if __name__ == '__main__':
    main()
