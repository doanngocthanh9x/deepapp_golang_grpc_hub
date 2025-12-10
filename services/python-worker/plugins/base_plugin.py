"""
Base Plugin Interface
All plugins must inherit from this class
"""

from abc import ABC, abstractmethod
from typing import Dict, Any, Optional


class BasePlugin(ABC):
    """
    Base class for all worker plugins
    
    Each plugin represents one capability that the worker can perform.
    Simply create a new .py file in the plugins/ directory and define a class
    that inherits from BasePlugin.
    """
    
    @property
    @abstractmethod
    def name(self) -> str:
        """Capability name (e.g., 'hello', 'process_image')"""
        pass
    
    @property
    @abstractmethod
    def description(self) -> str:
        """Human-readable description of what this plugin does"""
        pass
    
    @property
    def input_schema(self) -> str:
        """JSON schema for input validation (optional)"""
        return "{}"
    
    @property
    def output_schema(self) -> str:
        """JSON schema for output format (optional)"""
        return "{}"
    
    @property
    def http_method(self) -> str:
        """HTTP method for Web API endpoint (GET/POST/PUT/DELETE)"""
        return "POST"
    
    @property
    def accepts_file(self) -> bool:
        """Whether this capability accepts file upload"""
        return False
    
    @property
    def file_field_name(self) -> Optional[str]:
        """Field name for file upload (if accepts_file is True)"""
        return None
    
    @abstractmethod
    def execute(self, params: Dict[str, Any], worker_sdk=None) -> Dict[str, Any]:
        """
        Execute the plugin logic
        
        Args:
            params: Input parameters from the request
            worker_sdk: Reference to worker SDK for calling other workers
            
        Returns:
            Dict with the result data
        """
        pass
    
    def on_load(self):
        """Called when plugin is loaded (optional hook)"""
        pass
    
    def on_unload(self):
        """Called when plugin is unloaded (optional hook)"""
        pass
