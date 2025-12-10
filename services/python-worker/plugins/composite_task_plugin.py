"""
Composite Task Plugin - Demonstrates worker-to-worker communication
Calls Java worker to get file information
"""

from datetime import datetime
from plugins.base_plugin import BasePlugin


class CompositeTaskPlugin(BasePlugin):
    """Demo: Python processing + calls Java worker for file info"""
    
    @property
    def name(self) -> str:
        return "composite_task"
    
    @property
    def description(self) -> str:
        return "Demo: Python processing + calls Java worker for file info"
    
    @property
    def input_schema(self) -> str:
        return '{"type":"object","properties":{"file_path":{"type":"string"}}}'
    
    @property
    def output_schema(self) -> str:
        return '{"type":"object","properties":{"python_processing":{"type":"object"},"java_file_info":{"type":"object"}}}'
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        """Execute composite task with worker-to-worker call"""
        file_path = params.get('file_path', '/tmp/test.txt')
        
        # Step 1: Do Python processing
        python_result = {
            "processed_by": "python",
            "timestamp": datetime.now().isoformat(),
            "file_path": file_path
        }
        
        # Step 2: Call Java worker (if SDK provided)
        if worker_sdk is None:
            return {
                "python_processing": python_result,
                "error": "Worker SDK not available",
                "combined_status": "partial"
            }
        
        try:
            print(f"  → Calling Java worker for file info...")
            java_response = worker_sdk.call_worker(
                target_worker='java-simple-worker',
                capability='read_file_info',
                params={'filePath': file_path},
                timeout=30
            )
            
            return {
                "python_processing": python_result,
                "java_file_info": java_response,
                "combined_status": "success",
                "timestamp": datetime.now().isoformat()
            }
            
        except Exception as e:
            return {
                "python_processing": python_result,
                "java_call_error": str(e),
                "combined_status": "partial",
                "timestamp": datetime.now().isoformat()
            }
