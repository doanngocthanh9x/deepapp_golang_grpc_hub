"""
Hello Plugin - Simple greeting capability
"""

from datetime import datetime
from plugins.base_plugin import BasePlugin


class HelloPlugin(BasePlugin):
    """Simple hello world capability"""
    
    @property
    def name(self) -> str:
        return "hello"
    
    @property
    def description(self) -> str:
        return "Returns a hello message"
    
    @property
    def output_schema(self) -> str:
        return '{"type":"object","properties":{"message":{"type":"string"}}}'
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        return {
            "message": "Hello World from Python Worker! 🐍",
            "timestamp": datetime.now().isoformat(),
            "status": "success"
        }
