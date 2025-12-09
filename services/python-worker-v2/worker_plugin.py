"""
Dynamic Python Worker with Plugin System
Worker tự động load capabilities từ plugins/
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

# Import plugin system
from plugin_loader import PluginLoader
from decorators import get_registered_capabilities


class PluginWorker:
    """Worker with automatic plugin-based capability loading"""
    
    def __init__(self, worker_id, hub_address="localhost:50051", plugins_dir="plugins"):
        self.worker_id = worker_id
        self.hub_address = hub_address
        self.message_queue = Queue()
        self.stop_event = Event()
        self.capabilities = {}
        
        # Worker context to pass to handlers
        self.context = {
            "worker_id": worker_id,
            "worker_type": "python-plugin",
            "version": "2.0"
        }
        
        # Load plugins
        self._load_plugins(plugins_dir)
    
    def _load_plugins(self, plugins_dir):
        """Load all plugins from directory"""
        print(f"🔌 Loading plugins from: {plugins_dir}")
        loader = PluginLoader(plugins_dir)
        raw_capabilities = loader.load_plugins()
        
        # Convert to worker capabilities format
        for cap_name, cap_info in raw_capabilities.items():
            self.capabilities[cap_name] = {
                "name": cap_info["name"],
                "description": cap_info["description"],
                "input_schema": json.dumps(cap_info["input_schema"]),
                "output_schema": json.dumps(cap_info["output_schema"]),
                "handler": cap_info["handler"]
            }
        
        if not self.capabilities:
            print("⚠️  No capabilities loaded! Check plugins directory.")
    
    def create_registration_message(self):
        """Tạo message đăng ký với Hub"""
        registration = {
            "worker_id": self.worker_id,
            "worker_type": "python-plugin",
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
                "plugin_system": "enabled",
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
        print(f"✅ Worker {self.worker_id} registered with {len(self.capabilities)} capabilities", flush=True)
        
        # Then yield queued messages
        while not self.stop_event.is_set():
            try:
                msg = self.message_queue.get(timeout=1)
                print(f"📤 Sending response: {msg.id}", flush=True)
                yield msg
            except:
                continue
    
    def receive_loop(self, stub):
        """Receive and process incoming messages"""
        try:
            response_stream = stub.Connect(self.request_generator())
            print(f"✅ Connected to Hub, waiting for messages...", flush=True)
            
            for msg in response_stream:
                if self.stop_event.is_set():
                    break
                
                print(f"📩 Received message from {getattr(msg, 'from')}: {msg.id}", flush=True)
                print(f"   Type: {msg.type}, Content: {msg.content[:100]}...", flush=True)
                
                # Parse request
                try:
                    request_data = json.loads(msg.content)
                    capability = request_data.get("capability")
                    payload = request_data.get("payload", {})
                    
                    # Find handler
                    if capability in self.capabilities:
                        handler = self.capabilities[capability]["handler"]
                        
                        # Call handler with context
                        result = handler(self.context, payload)
                        
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
                        print(f"✅ Processed {capability} request, queued response", flush=True)
                    else:
                        print(f"❌ Unknown capability: {capability}")
                        # Send error response
                        error_msg = hub_pb2.Message(
                            id=msg.id,
                            type=hub_pb2.MessageType.RESPONSE,
                            content=json.dumps({"error": f"Unknown capability: {capability}"}),
                            timestamp=time.strftime("%Y-%m-%dT%H:%M:%SZ")
                        )
                        setattr(error_msg, 'from', self.worker_id)
                        setattr(error_msg, 'to', getattr(msg, 'from'))
                        self.message_queue.put(error_msg)
                        
                except json.JSONDecodeError as e:
                    print(f"❌ Invalid JSON: {e}")
                except Exception as e:
                    print(f"❌ Handler error: {e}")
                    import traceback
                    traceback.print_exc()
                    
        except grpc.RpcError as e:
            print(f"❌ gRPC error: {e}")
        except Exception as e:
            print(f"❌ Receive loop error: {e}")
    
    def run(self):
        """Start worker"""
        print(f"🚀 Starting Plugin Worker: {self.worker_id}")
        print(f"📡 Connecting to Hub: {self.hub_address}")
        print(f"🎯 Available capabilities: {', '.join(self.capabilities.keys())}")
        
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
    worker_id = os.getenv("WORKER_ID", f"py-plugin-worker-{int(time.time())}")
    plugins_dir = os.getenv("PLUGINS_DIR", "plugins")
    
    worker = PluginWorker(worker_id, hub_addr, plugins_dir)
    worker.run()
