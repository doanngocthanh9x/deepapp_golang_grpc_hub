"""
Dynamic Python Worker with Auto-Registration
Worker tự động đăng ký capabilities với Hub
"""

import grpc
import json
import time
import os
import sys
from threading import Thread, Event
from queue import Queue

# Import generated protobuf files
sys.path.append(os.path.join(os.path.dirname(__file__), '../../internal/proto'))
import hub_pb2
import hub_pb2_grpc


class DynamicWorker:
    """Worker with dynamic capability registration"""
    
    def __init__(self, worker_id, hub_address="localhost:50051"):
        self.worker_id = worker_id
        self.hub_address = hub_address
        self.message_queue = Queue()
        self.stop_event = Event()
        
        # Đăng ký các capabilities (handlers)
        self.capabilities = {
            "hello": {
                "name": "hello",
                "description": "Simple hello greeting service",
                "input_schema": json.dumps({
                    "type": "object",
                    "properties": {
                        "name": {"type": "string", "description": "Name to greet"}
                    }
                }),
                "output_schema": json.dumps({
                    "type": "object",
                    "properties": {
                        "message": {"type": "string"}
                    }
                }),
                "handler": self.handle_hello
            },
            "image_analysis": {
                "name": "image_analysis",
                "description": "Analyze image and extract metadata",
                "input_schema": json.dumps({
                    "type": "object",
                    "properties": {
                        "image_url": {"type": "string", "description": "URL of image to analyze"}
                    }
                }),
                "output_schema": json.dumps({
                    "type": "object",
                    "properties": {
                        "width": {"type": "integer"},
                        "height": {"type": "integer"},
                        "format": {"type": "string"}
                    }
                }),
                "handler": self.handle_image_analysis
            },
            "text_processing": {
                "name": "text_processing",
                "description": "Process and analyze text content",
                "input_schema": json.dumps({
                    "type": "object",
                    "properties": {
                        "text": {"type": "string", "description": "Text to process"},
                        "operation": {"type": "string", "enum": ["count", "uppercase", "lowercase"]}
                    }
                }),
                "output_schema": json.dumps({
                    "type": "object",
                    "properties": {
                        "result": {"type": "string"},
                        "length": {"type": "integer"}
                    }
                }),
                "handler": self.handle_text_processing
            }
        }
    
    def handle_hello(self, payload):
        """Handler for hello capability"""
        name = payload.get("name", "World")
        return {
            "message": f"Hello, {name}! From worker {self.worker_id}"
        }
    
    def handle_image_analysis(self, payload):
        """Handler for image analysis capability"""
        image_url = payload.get("image_url", "")
        # Mock analysis
        return {
            "width": 1920,
            "height": 1080,
            "format": "jpeg",
            "url": image_url,
            "analyzed_by": self.worker_id
        }
    
    def handle_text_processing(self, payload):
        """Handler for text processing capability"""
        text = payload.get("text", "")
        operation = payload.get("operation", "count")
        
        result = text
        if operation == "uppercase":
            result = text.upper()
        elif operation == "lowercase":
            result = text.lower()
        
        return {
            "result": result,
            "length": len(text),
            "operation": operation,
            "processed_by": self.worker_id
        }
    
    def create_registration_message(self):
        """Tạo message đăng ký với Hub"""
        registration = {
            "worker_id": self.worker_id,
            "worker_type": "python",
            "capabilities": [
                {
                    "name": cap["name"],
                    "description": cap["description"],
                    "input_schema": cap["input_schema"],
                    "output_schema": cap["output_schema"]
                }
                for cap in self.capabilities.values()
            ],
            "metadata": {
                "version": "2.0",
                "language": "python",
                "started_at": time.strftime("%Y-%m-%dT%H:%M:%SZ")
            }
        }
        
        msg = hub_pb2.Message(
            id=f"reg-{int(time.time() * 1000)}",
            type=hub_pb2.MessageType.REGISTER,
            content=json.dumps(registration),
            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ")
        )
        setattr(msg, 'from', self.worker_id)
        setattr(msg, 'to', 'hub')
        
        return msg
    
    def request_generator(self):
        """Generator yields messages to send"""
        # First message: registration
        yield self.create_registration_message()
        print(f"✅ Worker {self.worker_id} registered with {len(self.capabilities)} capabilities")
        
        # Then yield queued messages
        while not self.stop_event.is_set():
            try:
                msg = self.message_queue.get(timeout=1)
                yield msg
            except:
                continue
    
    def receive_loop(self, stub):
        """Receive and process incoming messages"""
        try:
            response_stream = stub.Connect(self.request_generator())
            
            for msg in response_stream:
                if self.stop_event.is_set():
                    break
                
                print(f"📩 Received message from {getattr(msg, 'from')}: {msg.id}")
                
                # Parse request
                try:
                    request_data = json.loads(msg.content)
                    capability = request_data.get("capability")
                    payload = request_data.get("payload", {})
                    
                    # Find handler
                    if capability in self.capabilities:
                        handler = self.capabilities[capability]["handler"]
                        result = handler(payload)
                        
                        # Send response
                        response_msg = hub_pb2.Message(
                            id=msg.id,  # Same ID for correlation
                            type=hub_pb2.MessageType.RESPONSE,
                            content=json.dumps(result),
                            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ")
                        )
                        setattr(response_msg, 'from', self.worker_id)
                        setattr(response_msg, 'to', getattr(msg, 'from'))
                        
                        self.message_queue.put(response_msg)
                        print(f"✅ Processed {capability} request")
                    else:
                        print(f"❌ Unknown capability: {capability}")
                        
                except json.JSONDecodeError as e:
                    print(f"❌ Invalid JSON: {e}")
                except Exception as e:
                    print(f"❌ Handler error: {e}")
                    
        except grpc.RpcError as e:
            print(f"❌ gRPC error: {e}")
        except Exception as e:
            print(f"❌ Receive loop error: {e}")
    
    def run(self):
        """Start worker"""
        print(f"🚀 Starting Dynamic Worker: {self.worker_id}")
        print(f"📡 Connecting to Hub: {self.hub_address}")
        print(f"🎯 Registered capabilities: {', '.join(self.capabilities.keys())}")
        
        channel = grpc.insecure_channel(self.hub_address)
        stub = hub_pb2_grpc.HubServiceStub(channel)
        
        try:
            self.receive_loop(stub)
        except KeyboardInterrupt:
            print(f"\n⚠️  Shutting down worker {self.worker_id}...")
            self.stop_event.set()
        finally:
            channel.close()


if __name__ == "__main__":
    hub_addr = os.getenv("HUB_ADDRESS", "localhost:50051")
    worker_id = os.getenv("WORKER_ID", f"py-worker-{int(time.time())}")
    
    worker = DynamicWorker(worker_id, hub_addr)
    worker.run()
