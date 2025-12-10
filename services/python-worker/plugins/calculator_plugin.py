"""
Calculator Plugin - Simple math operations
Demo: No dependencies, pure Python logic
"""

from plugins.base_plugin import BasePlugin


class CalculatorPlugin(BasePlugin):
    """Performs basic math operations"""
    
    @property
    def name(self) -> str:
        return "calculate"
    
    @property
    def description(self) -> str:
        return "Performs basic math operations (add, subtract, multiply, divide)"
    
    @property
    def input_schema(self) -> str:
        return '''{"type":"object","properties":{"operation":{"type":"string","enum":["add","subtract","multiply","divide"]},"a":{"type":"number"},"b":{"type":"number"}}}'''
    
    @property
    def output_schema(self) -> str:
        return '{"type":"object","properties":{"result":{"type":"number"},"operation":{"type":"string"}}}'
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        """Perform calculation"""
        operation = params.get('operation', 'add')
        a = params.get('a', 0)
        b = params.get('b', 0)
        
        if operation == 'add':
            result = a + b
        elif operation == 'subtract':
            result = a - b
        elif operation == 'multiply':
            result = a * b
        elif operation == 'divide':
            if b == 0:
                return {"error": "Division by zero", "status": "failed"}
            result = a / b
        else:
            return {"error": f"Unknown operation: {operation}", "status": "failed"}
        
        return {
            "result": result,
            "operation": operation,
            "a": a,
            "b": b,
            "status": "success"
        }
