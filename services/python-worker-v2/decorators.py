"""
Decorators for auto-registering worker capabilities
"""

import functools
from typing import Callable, Dict, Any


# Global registry for decorated handlers
_CAPABILITY_REGISTRY: Dict[str, Dict[str, Any]] = {}


def capability(
    name: str,
    description: str,
    input_schema: dict = None,
    output_schema: dict = None
):
    """
    Decorator to mark a method as a capability handler
    
    Usage:
        @capability(
            name="hello",
            description="Simple greeting service",
            input_schema={"type": "object", "properties": {"name": {"type": "string"}}},
            output_schema={"type": "object", "properties": {"message": {"type": "string"}}}
        )
        def handle_hello(self, payload):
            return {"message": f"Hello {payload.get('name', 'World')}"}
    """
    def decorator(func: Callable) -> Callable:
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)
        
        # Register capability
        _CAPABILITY_REGISTRY[name] = {
            "name": name,
            "description": description,
            "input_schema": input_schema or {},
            "output_schema": output_schema or {},
            "handler": func,
            "function_name": func.__name__
        }
        
        # Mark function as capability
        wrapper._is_capability = True
        wrapper._capability_name = name
        wrapper._capability_info = _CAPABILITY_REGISTRY[name]
        
        return wrapper
    
    return decorator


def get_registered_capabilities() -> Dict[str, Dict[str, Any]]:
    """Get all registered capabilities"""
    return _CAPABILITY_REGISTRY.copy()


def clear_registry():
    """Clear capability registry (useful for testing)"""
    _CAPABILITY_REGISTRY.clear()
