#!/usr/bin/env python3
"""
Example Worker using the WorkerSDK
Demonstrates how easy it is to create a worker with worker-to-worker communication
"""

import os
import sys

# Add SDK to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '../../../shared/worker-sdk/python'))

from worker_sdk import WorkerSDK


class ExampleWorker(WorkerSDK):
    """Example worker demonstrating SDK usage"""
    
    def register_capabilities(self):
        """Register this worker's capabilities"""
        
        # Simple hello capability
        self.add_capability(
            name="hello",
            handler=self.handle_hello,
            description="Returns a hello message",
            http_method="POST"
        )
        
        # Image analysis capability
        self.add_capability(
            name="analyze_image",
            handler=self.handle_analyze_image,
            description="Analyzes an uploaded image",
            http_method="POST",
            accepts_file=True,
            file_field_name="file"
        )
        
        # Composite task - demonstrates calling other workers
        self.add_capability(
            name="composite_task",
            handler=self.handle_composite_task,
            description="Demo: Calls Java worker for file info",
            http_method="POST",
            input_schema='{"type":"object","properties":{"file_path":{"type":"string"}}}',
            output_schema='{"type":"object","properties":{"python_processing":{"type":"object"},"java_file_info":{"type":"object"}}}'
        )
    
    def handle_hello(self, params: dict) -> dict:
        """Simple hello handler"""
        return {
            "message": "Hello from Python Worker! 🐍",
            "worker_id": self.worker_id,
            "status": "success"
        }
    
    def handle_analyze_image(self, params: dict) -> dict:
        """Analyze an image"""
        filename = params.get('filename', 'unknown')
        size = params.get('size', 0)
        
        return {
            "filename": filename,
            "size_bytes": size,
            "format": filename.split('.')[-1].upper() if '.' in filename else "UNKNOWN",
            "analysis": {
                "detected_objects": ["person", "car", "tree"],
                "confidence": 0.85
            },
            "status": "success"
        }
    
    def handle_composite_task(self, params: dict) -> dict:
        """
        Composite task that calls Java worker
        Demonstrates worker-to-worker communication
        """
        file_path = params.get('file_path', '/tmp/test.txt')
        
        # Step 1: Do local processing
        python_result = {
            "processed_by": "python",
            "file_path": file_path
        }
        
        # Step 2: Call Java worker for file info
        try:
            self.log(f"  → Calling Java worker for file info...")
            java_response = self.call_worker(
                target_worker='java-simple-worker',
                capability='read_file_info',
                params={'filePath': file_path},
                timeout=30
            )
            
            return {
                "python_processing": python_result,
                "java_file_info": java_response,
                "combined_status": "success",
                "worker_id": self.worker_id
            }
        
        except Exception as e:
            # Return partial result on error
            return {
                "python_processing": python_result,
                "java_call_error": str(e),
                "combined_status": "partial",
                "worker_id": self.worker_id
            }


def main():
    # Get configuration from environment
    worker_id = os.getenv('WORKER_ID', 'python-worker')
    hub_address = os.getenv('HUB_ADDRESS', 'localhost:50051')
    
    # Create and run worker
    worker = ExampleWorker(worker_id, hub_address)
    
    try:
        worker.run()
    except KeyboardInterrupt:
        print("\n\n✗ Shutting down...")
        worker.stop()


if __name__ == '__main__':
    main()
