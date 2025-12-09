"""
Basic text and greeting plugins
"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from decorators import capability


@capability(
    name="hello",
    description="Simple hello greeting service",
    input_schema={
        "type": "object",
        "properties": {
            "name": {"type": "string", "description": "Name to greet"}
        }
    },
    output_schema={
        "type": "object",
        "properties": {
            "message": {"type": "string"}
        }
    }
)
def handle_hello(worker_context, payload):
    """Say hello to someone"""
    name = payload.get("name", "World")
    worker_id = worker_context.get("worker_id", "unknown")
    return {
        "message": f"Hello, {name}! From worker {worker_id}"
    }


@capability(
    name="echo",
    description="Echo back the input message",
    input_schema={
        "type": "object",
        "properties": {
            "message": {"type": "string", "description": "Message to echo"}
        },
        "required": ["message"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "echo": {"type": "string"},
            "length": {"type": "integer"}
        }
    }
)
def handle_echo(worker_context, payload):
    """Echo input message"""
    message = payload.get("message", "")
    return {
        "echo": message,
        "length": len(message),
        "worker": worker_context.get("worker_id")
    }


@capability(
    name="text_transform",
    description="Transform text (uppercase, lowercase, title, reverse)",
    input_schema={
        "type": "object",
        "properties": {
            "text": {"type": "string", "description": "Text to transform"},
            "operation": {
                "type": "string",
                "enum": ["uppercase", "lowercase", "title", "reverse", "count"],
                "description": "Transformation operation"
            }
        },
        "required": ["text", "operation"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "result": {"type": "string"},
            "operation": {"type": "string"},
            "original_length": {"type": "integer"}
        }
    }
)
def handle_text_transform(worker_context, payload):
    """Transform text in various ways"""
    text = payload.get("text", "")
    operation = payload.get("operation", "uppercase")
    
    result = text
    if operation == "uppercase":
        result = text.upper()
    elif operation == "lowercase":
        result = text.lower()
    elif operation == "title":
        result = text.title()
    elif operation == "reverse":
        result = text[::-1]
    elif operation == "count":
        result = f"Characters: {len(text)}, Words: {len(text.split())}"
    
    return {
        "result": result,
        "operation": operation,
        "original_length": len(text),
        "processed_by": worker_context.get("worker_id")
    }
