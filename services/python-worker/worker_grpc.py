#!/usr/bin/env python3
"""
Python Worker - Real gRPC Client
Connects to gRPC Hub using compiled proto files
"""

import grpc
import json
import time
import sys
import os
import uuid
import threading
from datetime import datetime

# Import generated proto files
import hub_pb2
import hub_pb2_grpc


class PythonWorker:
    """Worker that connects to gRPC Hub via bidirectional streaming"""
    
    def __init__(self, worker_id='python-worker', hub_address='localhost:50051'):
        self.worker_id = worker_id
        self.hub_address = hub_address
        self.channel = None
        self.stub = None
        self.running = False
        self.pending_calls = {}  # Track pending worker-to-worker calls
        self.pending_lock = threading.Lock()  # Thread-safe access
        
    def handle_hello(self, msg):
        """Handle hello request"""
        msg_from = getattr(msg, 'from')
        print(f"  → Processing hello from {msg_from}")
        
        response_data = {
            "message": "Hello World from Python Worker! 🐍",
            "timestamp": datetime.now().isoformat(),
            "worker_id": self.worker_id,
            "status": "success"
        }
        
        return json.dumps(response_data)
    
    def handle_image_analysis(self, msg):
        """Handle image analysis request"""
        msg_from = getattr(msg, 'from')
        print(f"  → Processing image analysis from {msg_from}")
        
        try:
            content = json.loads(msg.content)
            filename = content.get('filename', 'unknown')
            size = content.get('size', 0)
            
            response_data = {
                "filename": filename,
                "size_bytes": size,
                "size_mb": round(size / (1024 * 1024), 2),
                "format": filename.split('.')[-1].upper() if '.' in filename else "UNKNOWN",
                "analysis": {
                    "detected_objects": ["person", "car", "tree"],
                    "confidence": 0.85,
                    "colors": ["blue", "green", "red"],
                    "dimensions": "estimated 1920x1080"
                },
                "processing_time_ms": 150,
                "worker_id": self.worker_id,
                "timestamp": datetime.now().isoformat(),
                "status": "success"
            }
            
            return json.dumps(response_data)
            
        except Exception as e:
            error_data = {
                "error": str(e),
                "status": "failed",
                "worker_id": self.worker_id
            }
            return json.dumps(error_data)
    
    def handle_composite_task(self, msg):
        """
        Demo capability: Calls Java worker's file info capability
        This demonstrates worker-to-worker communication
        """
        msg_from = getattr(msg, 'from')
        print(f"  → Processing composite task from {msg_from}")
        
        try:
            content = json.loads(msg.content)
            file_path = content.get('file_path', '/tmp/test.txt')
            
            # Step 1: Do Python processing
            python_result = {
                "processed_by": "python",
                "timestamp": datetime.now().isoformat(),
                "file_path": file_path
            }
            
            # Step 2: Call Java worker to get file info
            print(f"  → Calling Java worker for file info...")
            try:
                java_response = self.call_worker(
                    target_worker='java-simple-worker',
                    capability='read_file_info',
                    params={'filePath': file_path},  # Match Java's expected field name
                    timeout=10
                )
                
                # Combine results
                response_data = {
                    "python_processing": python_result,
                    "java_file_info": java_response,
                    "combined_status": "success",
                    "worker_id": self.worker_id,
                    "timestamp": datetime.now().isoformat()
                }
                
                return json.dumps(response_data)
                
            except Exception as e:
                # If Java worker call fails, still return partial result
                response_data = {
                    "python_processing": python_result,
                    "java_call_error": str(e),
                    "combined_status": "partial",
                    "worker_id": self.worker_id,
                    "timestamp": datetime.now().isoformat()
                }
                return json.dumps(response_data)
            
        except Exception as e:
            error_data = {
                "error": str(e),
                "status": "failed",
                "worker_id": self.worker_id
            }
            return json.dumps(error_data)
    
    def call_worker(self, target_worker, capability, params, timeout=30):
        """
        Call another worker's capability through Hub
        
        Args:
            target_worker: ID of the target worker (e.g., 'java-simple-worker')
            capability: Capability name to call (e.g., 'read_file_info')
            params: Dictionary of parameters to send
            timeout: Timeout in seconds (default 30)
        
        Returns:
            Dictionary with response data
        
        Raises:
            TimeoutError: If no response within timeout
            Exception: If error response received
        """
        if not hasattr(self, 'send_queue'):
            raise RuntimeError("Worker not connected. Call connect_and_run() first")
        
        request_id = str(uuid.uuid4())
        
        print(f"🔗 Calling worker '{target_worker}' capability '{capability}'")
        
        # Create worker call message
        call_msg = hub_pb2.Message(
            id=request_id,
            to=target_worker,
            channel=capability,
            content=json.dumps(params),
            timestamp=datetime.now().isoformat(),
            type=hub_pb2.WORKER_CALL  # Use new message type
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
            print(f"  → Registered pending call {request_id}")
            print(f"  → Total pending calls after register: {len(self.pending_calls)}")
            print(f"  → All pending call IDs: {list(self.pending_calls.keys())}")
        
        # Send the call
        self.send_queue.put(call_msg)
        print(f"  → Sent WORKER_CALL message with ID {request_id}")
        
        # Wait for response
        print(f"  → Waiting for response (timeout: {timeout}s)...")
        if response_event.wait(timeout=timeout):
            print(f"  → Response event received for {request_id}")
            with self.pending_lock:
                removed = self.pending_calls.pop(request_id, None)
                print(f"  → Removed pending call {request_id}: {removed is not None}")
            
            if response_data['error']:
                raise Exception(response_data['error'])
            
            return response_data['response']
        else:
            # Timeout
            with self.pending_lock:
                self.pending_calls.pop(request_id, None)
            raise TimeoutError(f"No response from {target_worker} after {timeout}s")
    
    def _handle_worker_call_response(self, msg):
        """Handle response from worker-to-worker call"""
        # Get request ID from metadata to match with pending call
        request_id = msg.metadata.get('request_id', '')
        
        print(f"  → Checking response for request_id: {request_id}")
        print(f"     Message ID: {msg.id}, From: {getattr(msg, 'from', 'unknown')}, Type: {msg.type}")
        print(f"     Message metadata: {dict(msg.metadata)}")
        
        with self.pending_lock:
            print(f"     Current pending calls: {list(self.pending_calls.keys())}")
            if request_id and request_id in self.pending_calls:
                # Found matching pending call
                call_info = self.pending_calls[request_id]
                try:
                    response_content = json.loads(msg.content)
                    call_info['data']['response'] = response_content
                    call_info['event'].set()
                    print(f"  ✓ Matched and completed pending call {request_id}")
                except Exception as e:
                    call_info['data']['error'] = f"Failed to parse response: {e}"
                    call_info['event'].set()
                    print(f"  ✗ Error parsing response: {e}")
            else:
                # No matching pending call - might be a regular response
                print(f"  ⚠️  No pending call found for request_id: {request_id}")
                print(f"     Available pending calls: {list(self.pending_calls.keys())}")
    
    def process_message(self, msg):
        """Process incoming message and return response content"""
        # Use channel field directly (simple and working)
        channel = msg.channel
        
        print(f"  → Processing channel: {channel}")
        
        if channel == 'hello':
            return self.handle_hello(msg)
        elif channel == 'analyze_image':
            return self.handle_image_analysis(msg)
        elif channel == 'composite_task':
            return self.handle_composite_task(msg)
        else:
            error = {
                "error": f"Unknown request type: {channel}",
                "status": "failed"
            }
            return json.dumps(error)
    
    def create_response(self, request_msg, content):
        """Create response message"""
        return hub_pb2.Message(
            id=f"resp-{int(time.time() * 1000000)}",
            from_=self.worker_id,
            to=request_msg.from_,
            channel=request_msg.channel,
            content=content,
            timestamp=datetime.now().isoformat(),
            type=hub_pb2.DIRECT
        )
    
    def message_generator(self, stream):
        """Generate messages to send to hub"""
        # First message: register with worker ID and capabilities
        capabilities = [
            {
                "name": "hello",
                "description": "Simple hello world response",
                "input_schema": "{}",
                "output_schema": "{\"message\": \"string\"}",
                "http_method": "GET",
                "accepts_file": False
            },
            {
                "name": "analyze_image",
                "description": "Analyze uploaded image",
                "input_schema": "{\"image\": \"file\"}",
                "output_schema": "{\"analysis\": \"string\", \"dimensions\": \"object\"}",
                "http_method": "POST",
                "accepts_file": True,
                "file_field_name": "image"
            }
        ]
        
        registration_data = {
            "worker_id": self.worker_id,
            "worker_type": "python",
            "capabilities": capabilities,
            "metadata": {
                "version": "1.0.0",
                "description": "Python worker with image analysis"
            }
        }
        
        register_msg = hub_pb2.Message(
            id=f"register-{int(time.time() * 1000000)}",
            from_=self.worker_id,
            to="hub",
            channel="system",
            content=json.dumps(registration_data),
            timestamp=datetime.now().isoformat(),
            type=hub_pb2.REGISTER
        )
        
        print(f"📤 Sending registration message")
        yield register_msg
        
        # Process incoming messages and send responses
        try:
            for response_msg in stream:
                if response_msg:
                    yield response_msg
        except Exception as e:
            print(f"✗ Generator error: {e}")
    
    def connect_and_run(self):
        """Connect to gRPC Hub and start processing"""
        print(f"🐍 Python Worker (gRPC)")
        print(f"=" * 50)
        print(f"Worker ID: {self.worker_id}")
        print(f"Hub Address: {self.hub_address}")
        print(f"=" * 50)
        
        try:
            # Create channel
            print(f"Connecting to gRPC Hub...")
            self.channel = grpc.insecure_channel(self.hub_address)
            
            # Wait for channel to be ready
            grpc.channel_ready_future(self.channel).result(timeout=10)
            print(f"✓ Connected to Hub")
            
            # Create stub
            self.stub = hub_pb2_grpc.HubServiceStub(self.channel)
            
            # Create message queue for sending
            import queue
            import threading
            send_queue = queue.Queue()
            self.send_queue = send_queue  # Make accessible for call_worker method
            
            # Generator function for sending messages
            def request_generator():
                try:
                    # Send registration message first
                    capabilities = [
                        {
                            "name": "hello",
                            "description": "Returns a hello message",
                            "input_schema": "{}",
                            "output_schema": "{\"type\":\"object\",\"properties\":{\"message\":{\"type\":\"string\"}}}",
                            "http_method": "POST",
                            "accepts_file": False
                        },
                        {
                            "name": "analyze_image",
                            "description": "Analyzes an image and returns information",
                            "input_schema": "{\"type\":\"object\",\"properties\":{\"file\":{\"type\":\"string\",\"format\":\"binary\"}}}",
                            "output_schema": "{\"type\":\"object\",\"properties\":{\"result\":{\"type\":\"string\"}}}",
                            "http_method": "POST",
                            "accepts_file": True,
                            "file_field_name": "file"
                        },
                        {
                            "name": "composite_task",
                            "description": "Demo: Python processing + calls Java worker for file info",
                            "input_schema": "{\"type\":\"object\",\"properties\":{\"file_path\":{\"type\":\"string\"}}}",
                            "output_schema": "{\"type\":\"object\",\"properties\":{\"python_processing\":{\"type\":\"object\"},\"java_file_info\":{\"type\":\"object\"}}}",
                            "http_method": "POST",
                            "accepts_file": False
                        }
                    ]
                    
                    registration_data = {
                        "worker_id": self.worker_id,
                        "worker_type": "python",
                        "capabilities": capabilities,
                        "metadata": {
                            "version": "1.0.0",
                            "description": "Python worker with image analysis"
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
                    # Set 'from' field using setattr since 'from' is a Python keyword
                    setattr(register_msg, 'from', self.worker_id)
                    yield register_msg
                    print(f"📤 Sent registration message")
                    
                    # Keep generator alive by continuously yielding messages
                    # Use blocking get() to wait for messages instead of timeout
                    while self.running:
                        try:
                            # Block until we have a message to send
                            # This keeps the generator alive without busy waiting
                            msg = send_queue.get(block=True, timeout=1.0)
                            yield msg
                        except queue.Empty:
                            # No message within timeout, continue loop
                            # This is normal - keeps generator alive
                            continue
                        except Exception as e:
                            print(f"Error in generator: {e}")
                            break
                            
                except Exception as e:
                    print(f"Generator error: {e}")
                    import traceback
                    traceback.print_exc()
                finally:
                    print("Generator exiting...")
            
            # Start bidirectional streaming
            print(f"📡 Starting bidirectional stream...")
            response_stream = self.stub.Connect(request_generator())
            
            print(f"✓ Registered with Hub as '{self.worker_id}'")
            print(f"📨 Listening for requests...\n")
            
            self.running = True
            
            # Thread for receiving messages
            def receive_loop():
                try:
                    for msg in response_stream:
                        if not self.running:
                            break
                            
                        try:
                            msg_from = getattr(msg, 'from')  # Get 'from' field
                            msg_type = msg.type
                            
                            print(f"📬 Received message:")
                            print(f"   ID: {msg.id}")
                            print(f"   From: {msg_from}")
                            print(f"   Type: {msg_type}")
                            print(f"   Channel: {msg.channel}")
                            
                            # Handle different message types
                            if msg_type == hub_pb2.RESPONSE:
                                # This is a response (possibly from worker-to-worker call)
                                print(f"   → Response message")
                                self._handle_worker_call_response(msg)
                                # Don't send another response for RESPONSE messages
                                continue
                                
                            elif msg_type == hub_pb2.WORKER_CALL:
                                # Another worker is calling us
                                print(f"   → Worker call from {msg_from}")
                                # Process and send response
                                response_content = self.process_message(msg)
                                
                                response_msg = hub_pb2.Message(
                                    id=f"resp-{int(time.time() * 1000000)}",
                                    to=msg_from,
                                    channel=msg.channel,
                                    content=response_content,
                                    timestamp=datetime.now().isoformat(),
                                    type=hub_pb2.RESPONSE  # Mark as RESPONSE
                                )
                                setattr(response_msg, 'from', self.worker_id)
                                send_queue.put(response_msg)
                                print(f"   ✓ Queued response for worker call\n")
                                
                            else:
                                # Regular REQUEST or other message types
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
                                print(f"   ✓ Queued response for {msg_from}\n")
                            
                        except Exception as e:
                            print(f"✗ Error processing message: {e}")
                            import traceback
                            traceback.print_exc()
                            
                except grpc.RpcError as e:
                    if e.code() == grpc.StatusCode.CANCELLED:
                        print("Stream cancelled")
                    else:
                        print(f"✗ gRPC Error: {e.code()} - {e.details()}")
                except Exception as e:
                    print(f"✗ Receive loop error: {e}")
                    import traceback
                    traceback.print_exc()
                finally:
                    self.running = False
                    print("Receive loop exiting...")
            
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
            import traceback
            traceback.print_exc()
        finally:
            self.running = False
            if self.channel:
                self.channel.close()
                print("\n✗ Disconnected from Hub")
    
    def stop(self):
        """Stop the worker"""
        self.running = False


def main():
    # Get configuration from environment variables
    worker_id = os.getenv('WORKER_ID', 'python-worker')
    hub_address = os.getenv('HUB_ADDRESS', 'localhost:50051')
    
    worker = PythonWorker(worker_id=worker_id, hub_address=hub_address)
    
    try:
        worker.connect_and_run()
    except KeyboardInterrupt:
        print("\n\n✗ Shutting down...")
        worker.stop()


if __name__ == '__main__':
    main()
