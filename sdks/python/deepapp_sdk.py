"""
DeepApp gRPC Hub - Python Worker SDK

This SDK provides a base Worker class for creating workers that connect
to the DeepApp gRPC Hub system.
"""

import grpc
import json
import time
import threading
import queue
import os
import base64
from datetime import datetime

# Import generated proto files (you need to generate these first)
# Run: python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. ../../proto/hub.proto
try:
    import hub_pb2
    import hub_pb2_grpc
except ImportError:
    print("Error: Proto files not found. Please generate them first:")
    print("python -m grpc_tools.protoc -I../../proto --python_out=. --grpc_python_out=. ../../proto/hub.proto")
    raise


class Worker:
    """Base worker class for DeepApp gRPC Hub"""

    def __init__(self, worker_id=None, hub_address='localhost:50051'):
        self.worker_id = worker_id or f'python-worker-{int(time.time())}'
        self.hub_address = hub_address
        self.channel = None
        self.stub = None
        self.running = False
        self.send_queue = queue.Queue()

    def get_capabilities(self):
        """Override this method to define your worker's capabilities"""
        return []

    def start(self):
        """Connect to gRPC Hub and start processing"""
        print(f"🐍 Starting Python Worker: {self.worker_id}")
        print(f"📡 Connecting to Hub at: {self.hub_address}")

        try:
            # Create gRPC channel
            self.channel = grpc.insecure_channel(self.hub_address)

            # Wait for channel to be ready
            grpc.channel_ready_future(self.channel).result(timeout=10)
            print("✓ Connected to Hub")

            # Create stub
            self.stub = hub_pb2_grpc.HubServiceStub(self.channel)

            # Start bidirectional stream
            print("📡 Starting bidirectional stream...")
            response_stream = self.stub.Connect(self._request_generator())

            print(f"✓ Registered with Hub as '{self.worker_id}'")
            print("📨 Listening for requests...\n")

            self.running = True

            # Thread for receiving messages
            receive_thread = threading.Thread(target=self._receive_loop, args=(response_stream,))
            receive_thread.daemon = True
            receive_thread.start()

            # Keep main thread alive
            print("Worker running. Press Ctrl+C to stop.\n")
            while self.running:
                time.sleep(1)

        except grpc.RpcError as e:
            print(f"✗ gRPC Error: {e.code()} - {e.details()}")
        except Exception as e:
            print(f"✗ Connection error: {e}")
            import traceback
            traceback.print_exc()
        finally:
            self.running = False
            if self.channel:
                self.channel.close()
                print("\n✗ Disconnected from Hub")

    def stop(self):
        """Stop the worker"""
        print("🛑 Stopping worker...")
        self.running = False

    def _request_generator(self):
        """Generate messages to send to hub"""
        try:
            # Send registration message first
            capabilities = self.get_capabilities()

            registration_data = {
                'worker_id': self.worker_id,
                'worker_type': 'python-sdk',
                'capabilities': capabilities,
                'metadata': {
                    'version': '1.0.0',
                    'description': 'Python SDK Worker',
                    'sdk': 'python-sdk'
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

            yield register_msg
            print("📤 Sent registration message")

            # Keep generator alive by continuously yielding messages
            while self.running:
                try:
                    # Block until we have a message to send
                    msg = self.send_queue.get(block=True, timeout=1.0)
                    yield msg
                except queue.Empty:
                    # No message within timeout, continue loop
                    continue
                except Exception as e:
                    print(f"Error in generator: {e}")
                    break

        except Exception as e:
            print(f"Generator error: {e}")
            import traceback
            traceback.print_exc()

    def _receive_loop(self, response_stream):
        """Handle incoming messages"""
        try:
            for msg in response_stream:
                if not self.running:
                    break

                try:
                    msg_from = getattr(msg, 'from')
                    print(f"📬 Received request:")
                    print(f"   ID: {msg.id}")
                    print(f"   From: {msg_from}")
                    print(f"   Channel: {msg.channel}")

                    # Process the message
                    response_content = self._process_message(msg)

                    # Create response
                    response_msg = hub_pb2.Message(
                        id=f"resp-{int(time.time() * 1000000)}",
                        to=msg_from,
                        channel=msg.channel,
                        content=json.dumps(response_content),
                        timestamp=datetime.now().isoformat(),
                        type=hub_pb2.DIRECT
                    )
                    setattr(response_msg, 'from', self.worker_id)

                    # Put response in send queue
                    self.send_queue.put(response_msg)
                    print("   ✓ Queued response\n")

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

    def _process_message(self, msg):
        """Process incoming message and return response"""
        capability = msg.channel

        print(f"  → Processing capability: {capability}")

        # Find handler method
        handler_method = getattr(self, f'handle_{capability}', None)

        if handler_method and callable(handler_method):
            try:
                return handler_method(msg)
            except Exception as e:
                print(f"✗ Error in handler {capability}: {e}")
                return {
                    'error': str(e),
                    'status': 'failed',
                    'capability': capability
                }
        else:
            error = {
                'error': f'Unknown capability: {capability}',
                'status': 'failed'
            }
            return error


class SimpleWorker(Worker):
    """Example worker with basic capabilities"""

    def get_capabilities(self):
        return [
            {
                'name': 'hello',
                'description': 'Returns a hello message',
                'input_schema': '{}',
                'output_schema': '{"type":"object","properties":{"message":{"type":"string"},"timestamp":{"type":"string"},"worker_id":{"type":"string"}}}',
                'http_method': 'GET',
                'accepts_file': False
            },
            {
                'name': 'echo',
                'description': 'Echoes back the input message',
                'input_schema': '{"type":"object","properties":{"message":{"type":"string"}}}',
                'output_schema': '{"type":"object","properties":{"echo":{"type":"string"},"timestamp":{"type":"string"}}}',
                'http_method': 'POST',
                'accepts_file': False
            },
            {
                'name': 'process_file',
                'description': 'Process an uploaded file',
                'input_schema': '{"type":"object","properties":{"file":{"type":"string","format":"binary"},"filename":{"type":"string"}}}',
                'output_schema': '{"type":"object","properties":{"filename":{"type":"string"},"size":{"type":"number"},"processed":{"type":"boolean"},"timestamp":{"type":"string"}}}',
                'http_method': 'POST',
                'accepts_file': True,
                'file_field_name': 'file'
            }
        ]

    def handle_hello(self, message):
        """Handle hello capability"""
        return {
            'message': 'Hello World from Python SDK Worker! 🐍',
            'timestamp': datetime.now().isoformat(),
            'worker_id': self.worker_id,
            'status': 'success'
        }

    def handle_echo(self, message):
        """Handle echo capability"""
        try:
            content = json.loads(message.content)
            input_message = content.get('message', 'No message provided')

            return {
                'echo': input_message,
                'timestamp': datetime.now().isoformat(),
                'status': 'success'
            }
        except Exception as e:
            return {
                'error': str(e),
                'status': 'failed'
            }

    def handle_process_file(self, message):
        """Handle file processing capability"""
        try:
            content = json.loads(message.content)
            filename = content.get('filename', 'unknown')
            file_data = content.get('file')

            if not file_data:
                return {
                    'error': 'No file data provided',
                    'status': 'failed'
                }

            # Decode base64 file data
            file_bytes = base64.b64decode(file_data)
            file_size = len(file_bytes)

            # Simulate file processing
            print(f"📁 Processing file: {filename} ({file_size} bytes)")

            # Here you would do actual file processing
            # For example: image analysis, text extraction, etc.

            return {
                'filename': filename,
                'size': file_size,
                'processed': True,
                'result': 'File processed successfully',
                'timestamp': datetime.now().isoformat(),
                'status': 'success'
            }
        except Exception as e:
            print(f"Error processing file: {e}")
            return {
                'error': str(e),
                'status': 'failed'
            }


def main():
    # Get configuration from environment variables
    worker_id = os.getenv('WORKER_ID')
    hub_address = os.getenv('HUB_ADDRESS', 'localhost:50051')

    worker = SimpleWorker(worker_id=worker_id, hub_address=hub_address)

    try:
        worker.start()
    except KeyboardInterrupt:
        print("\n\n✗ Shutting down...")
        worker.stop()


if __name__ == '__main__':
    main()