"""
Math and calculation plugins
"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from decorators import capability


@capability(
    name="calculator",
    description="Perform basic math calculations",
    input_schema={
        "type": "object",
        "properties": {
            "operation": {
                "type": "string",
                "enum": ["add", "subtract", "multiply", "divide", "power", "sqrt"],
                "description": "Math operation"
            },
            "a": {"type": "number", "description": "First operand"},
            "b": {"type": "number", "description": "Second operand (not required for sqrt)"}
        },
        "required": ["operation", "a"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "result": {"type": "number"},
            "operation": {"type": "string"}
        }
    }
)
def handle_calculator(worker_context, payload):
    """Perform mathematical calculations"""
    operation = payload.get("operation")
    a = float(payload.get("a", 0))
    b = float(payload.get("b", 0))
    
    result = 0
    if operation == "add":
        result = a + b
    elif operation == "subtract":
        result = a - b
    elif operation == "multiply":
        result = a * b
    elif operation == "divide":
        if b == 0:
            return {"error": "Division by zero"}
        result = a / b
    elif operation == "power":
        result = a ** b
    elif operation == "sqrt":
        result = a ** 0.5
    
    return {
        "result": result,
        "operation": operation,
        "operands": {"a": a, "b": b} if operation != "sqrt" else {"a": a},
        "calculated_by": worker_context.get("worker_id")
    }


@capability(
    name="statistics",
    description="Calculate statistics for a list of numbers",
    input_schema={
        "type": "object",
        "properties": {
            "numbers": {
                "type": "array",
                "items": {"type": "number"},
                "description": "List of numbers"
            }
        },
        "required": ["numbers"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "count": {"type": "integer"},
            "sum": {"type": "number"},
            "mean": {"type": "number"},
            "min": {"type": "number"},
            "max": {"type": "number"}
        }
    }
)
def handle_statistics(worker_context, payload):
    """Calculate statistics from a list of numbers"""
    numbers = payload.get("numbers", [])
    
    if not numbers:
        return {"error": "Empty list"}
    
    return {
        "count": len(numbers),
        "sum": sum(numbers),
        "mean": sum(numbers) / len(numbers),
        "min": min(numbers),
        "max": max(numbers),
        "median": sorted(numbers)[len(numbers) // 2],
        "processed_by": worker_context.get("worker_id")
    }
