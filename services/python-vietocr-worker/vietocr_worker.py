#!/usr/bin/env python3
"""
VietOCR Worker - gRPC Client
Kết nối tới DeepApp Hub và xử lý OCR requests
"""

import grpc
import json
import time
import sys
import os
import uuid
import threading
from datetime import datetime

# Import generated proto files from parent
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../python-worker'))
import hub_pb2
import hub_pb2_grpc

# Import VietOCR ONNX inference
from vietocr_onnx import VietOCR_ONNX


class VietOCRWorker:
    """VietOCR Worker với gRPC Hub integration"""
    
    def __init__(
        self, 
        worker_id='python-vietocr-worker',
        hub_address='localhost:50051',
        encoder_path=None,
        decoder_path=None,
        use_gpu=False
    ):
        self.worker_id = worker_id
        self.hub_address = hub_address
        self.channel = None
        self.stub = None
        self.running = False
        self.send_queue = None
        
        # Initialize VietOCR (chỉ nếu có models)
        if encoder_path and decoder_path and os.path.exists(encoder_path) and os.path.exists(decoder_path):
            print(f"🔧 Initializing VietOCR ONNX...")
            self.ocr = VietOCR_ONNX(
                encoder_path=encoder_path,
                decoder_path=decoder_path,
                use_gpu=use_gpu
            )
            print(f"✓ VietOCR ready")
        else:
            print(f"⚠️  VietOCR models not found - running in demo mode")
            self.ocr = None
    
    def handle_ocr_detect(self, msg):
        """Handle single image OCR request"""
        msg_from = getattr(msg, 'from')
        print(f"  → Processing OCR detect from {msg_from}")
        
        try:
            start_time = time.time()
            content = json.loads(msg.content)
            image_data = content.get('image')
            
            if not image_data:
                raise ValueError("Missing 'image' field")
            
            # Run OCR
            if self.ocr:
                text, confidence = self.ocr.predict(image_data, return_prob=True)
            else:
                # Demo mode
                text = "Demo: VietOCR text recognition (models not loaded)"
                confidence = 0.95
            
            processing_time = (time.time() - start_time) * 1000
            
            response_data = {
                "text": text,
                "confidence": float(confidence),
                "processing_time_ms": round(processing_time, 2),
                "worker_id": self.worker_id,
                "status": "success"
            }
            
            print(f"  ✓ OCR result: '{text[:50]}...' (conf: {confidence:.3f}, time: {processing_time:.1f}ms)")
            return json.dumps(response_data)
            
        except Exception as e:
            print(f"  ✗ OCR error: {e}")
            return json.dumps({
                "status": "error",
                "error": str(e),
                "worker_id": self.worker_id
            })
    
    def handle_ocr_batch(self, msg):
        """Handle batch OCR request"""
        msg_from = getattr(msg, 'from')
        print(f"  → Processing OCR batch from {msg_from}")
        
        try:
            start_time = time.time()
            content = json.loads(msg.content)
            images = content.get('images', [])
            
            if not images:
                raise ValueError("Missing 'images' field")
            
            print(f"  📦 Processing {len(images)} images...")
            
            results = []
            for idx, image_data in enumerate(images):
                try:
                    if self.ocr:
                        text, confidence = self.ocr.predict(image_data, return_prob=True)
                    else:
                        text = f"Demo text {idx+1}"
                        confidence = 0.90
                    
                    results.append({
                        "text": text,
                        "confidence": float(confidence),
                        "index": idx
                    })
                except Exception as e:
                    results.append({
                        "text": "",
                        "confidence": 0.0,
                        "index": idx,
                        "error": str(e)
                    })
            
            processing_time = (time.time() - start_time) * 1000
            
            response_data = {
                "results": results,
                "total_images": len(images),
                "successful": sum(1 for r in results if not r.get('error')),
                "total_processing_time_ms": round(processing_time, 2),
                "worker_id": self.worker_id,
                "status": "success"
            }
            
            print(f"  ✓ Batch complete: {len(results)} images in {processing_time:.1f}ms")
            return json.dumps(response_data)
            
        except Exception as e:
            print(f"  ✗ Batch error: {e}")
            return json.dumps({
                "status": "error",
                "error": str(e),
                "worker_id": self.worker_id
            })
    
    def process_message(self, msg):
        """Process incoming message based on capability"""
        capability = msg.metadata.get('capability', msg.channel)
        
        print(f"📨 Processing capability: {capability}")
        
        handlers = {
            'ocr_detect': self.handle_ocr_detect,
            'ocr_batch': self.handle_ocr_batch,
        }
        
        handler = handlers.get(capability)
        if handler:
            return handler(msg)
        else:
            return json.dumps({
                "status": "error",
                "error": f"Unknown capability: {capability}",
                "worker_id": self.worker_id
            })
    
    def connect_and_run(self):
        """Connect to Hub and start processing"""
        print(f"🐍 VietOCR Worker (Python + ONNX)")
        print(f"=" * 60)
        print(f"Worker ID: {self.worker_id}")
        print(f"Hub Address: {self.hub_address}")
        print(f"=" * 60)
        
        try:
            # Connect to Hub
            print(f"Connecting to gRPC Hub...")
            self.channel = grpc.insecure_channel(self.hub_address)
            grpc.channel_ready_future(self.channel).result(timeout=10)
            print(f"✓ Connected to Hub")
            
            # Create stub
            self.stub = hub_pb2_grpc.HubServiceStub(self.channel)
            
            # Create send queue
            import queue
            send_queue = queue.Queue()
            self.send_queue = send_queue
            
            # Request generator
            def request_generator():
                try:
                    # Send registration
                    capabilities = [
                        {
                            "name": "ocr_detect",
                            "description": "OCR nhận diện text từ ảnh (Vietnamese + English)",
                            "input_schema": json.dumps({
                                "type": "object",
                                "properties": {
                                    "image": {"type": "string", "description": "Base64 encoded image"}
                                },
                                "required": ["image"]
                            }),
                            "output_schema": json.dumps({
                                "type": "object",
                                "properties": {
                                    "text": {"type": "string"},
                                    "confidence": {"type": "number"},
                                    "processing_time_ms": {"type": "number"}
                                }
                            }),
                            "http_method": "POST",
                            "accepts_file": True,
                            "file_field_name": "image"
                        },
                        {
                            "name": "ocr_batch",
                            "description": "Batch OCR processing cho nhiều ảnh",
                            "input_schema": json.dumps({
                                "type": "object",
                                "properties": {
                                    "images": {
                                        "type": "array",
                                        "items": {"type": "string"}
                                    }
                                },
                                "required": ["images"]
                            }),
                            "output_schema": json.dumps({
                                "type": "object",
                                "properties": {
                                    "results": {"type": "array"},
                                    "total_processing_time_ms": {"type": "number"}
                                }
                            }),
                            "http_method": "POST",
                            "accepts_file": False
                        }
                    ]
                    
                    registration_data = {
                        "worker_id": self.worker_id,
                        "worker_type": "python-vietocr",
                        "capabilities": capabilities,
                        "metadata": {
                            "version": "1.0.0",
                            "description": "VietOCR ONNX Worker - Vietnamese OCR",
                            "language": "Vietnamese + English",
                            "engine": "ONNX Runtime"
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
                    print(f"📤 Sent registration")
                    
                    # Keep generator alive
                    while self.running:
                        try:
                            msg = send_queue.get(block=True, timeout=1.0)
                            yield msg
                        except:
                            continue
                            
                except Exception as e:
                    print(f"✗ Generator error: {e}")
            
            # Start bidirectional streaming
            print(f"📡 Starting bidirectional stream...")
            response_stream = self.stub.Connect(request_generator())
            
            print(f"✓ Registered with Hub as '{self.worker_id}'")
            print(f"📨 Listening for OCR requests...\n")
            
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
                            print(f"   ID: {msg.id}")
                            print(f"   From: {msg_from}")
                            print(f"   Type: {msg_type}")
                            print(f"   Channel: {msg.channel}")
                            
                            # Process request
                            if msg_type == hub_pb2.REQUEST or msg_type == hub_pb2.WORKER_CALL:
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
                            
                except Exception as e:
                    print(f"✗ Receive loop error: {e}")
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
        finally:
            self.running = False
            if self.channel:
                self.channel.close()
    
    def stop(self):
        """Stop the worker"""
        self.running = False


def main():
    worker_id = os.getenv('WORKER_ID', 'python-vietocr-worker')
    hub_address = os.getenv('HUB_ADDRESS', 'localhost:50051')
    encoder_path = os.getenv('ENCODER_PATH', '/app/models/transformer_encoder.onnx')
    decoder_path = os.getenv('DECODER_PATH', '/app/models/transformer_decoder.onnx')
    use_gpu = os.getenv('USE_GPU', 'false').lower() == 'true'
    
    worker = VietOCRWorker(
        worker_id=worker_id,
        hub_address=hub_address,
        encoder_path=encoder_path,
        decoder_path=decoder_path,
        use_gpu=use_gpu
    )
    
    try:
        worker.connect_and_run()
    except KeyboardInterrupt:
        print("\n\n✗ Shutting down...")
        worker.stop()


if __name__ == '__main__':
    main()
