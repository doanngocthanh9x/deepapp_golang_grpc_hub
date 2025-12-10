"""
Plugin Manager - Auto-discovers and loads all plugins
"""

import os
import sys
import importlib
import inspect
from typing import Dict, List, Type
from plugins.base_plugin import BasePlugin


class PluginManager:
    """
    Auto-discovers and manages worker plugins
    
    Scans the plugins/ directory for Python files and automatically
    loads all classes that inherit from BasePlugin.
    """
    
    def __init__(self, plugins_dir: str = None):
        self.plugins_dir = plugins_dir or os.path.join(os.path.dirname(__file__), 'plugins')
        self.plugins: Dict[str, BasePlugin] = {}
        self.plugin_classes: Dict[str, Type[BasePlugin]] = {}
        
    def discover_plugins(self) -> List[str]:
        """
        Auto-discover all plugin files in plugins/ directory
        
        Returns:
            List of discovered plugin module names
        """
        if not os.path.exists(self.plugins_dir):
            print(f"⚠️  Plugins directory not found: {self.plugins_dir}")
            return []
        
        discovered = []
        
        # Add plugins directory to Python path
        if self.plugins_dir not in sys.path:
            sys.path.insert(0, os.path.dirname(self.plugins_dir))
        
        for filename in os.listdir(self.plugins_dir):
            if filename.endswith('_plugin.py') and not filename.startswith('_'):
                module_name = filename[:-3]  # Remove .py
                discovered.append(module_name)
                
        return discovered
    
    def load_plugin_module(self, module_name: str) -> List[Type[BasePlugin]]:
        """
        Load a plugin module and extract all BasePlugin classes
        
        Args:
            module_name: Name of the module (without .py)
            
        Returns:
            List of plugin classes found in the module
        """
        try:
            # Import the module
            module_path = f"plugins.{module_name}"
            module = importlib.import_module(module_path)
            
            # Find all classes that inherit from BasePlugin
            plugin_classes = []
            for name, obj in inspect.getmembers(module, inspect.isclass):
                if (issubclass(obj, BasePlugin) and 
                    obj is not BasePlugin and 
                    obj.__module__ == module_path):
                    plugin_classes.append(obj)
            
            return plugin_classes
            
        except Exception as e:
            print(f"✗ Error loading plugin module {module_name}: {e}")
            import traceback
            traceback.print_exc()
            return []
    
    def load_all_plugins(self) -> Dict[str, BasePlugin]:
        """
        Auto-discover and load all plugins
        
        Returns:
            Dictionary mapping capability names to plugin instances
        """
        print(f"\n🔌 Auto-discovering plugins from: {self.plugins_dir}")
        
        discovered_modules = self.discover_plugins()
        print(f"📦 Found {len(discovered_modules)} plugin modules: {discovered_modules}")
        
        loaded_count = 0
        for module_name in discovered_modules:
            plugin_classes = self.load_plugin_module(module_name)
            
            for plugin_class in plugin_classes:
                try:
                    # Instantiate the plugin
                    plugin_instance = plugin_class()
                    capability_name = plugin_instance.name
                    
                    # Register the plugin
                    self.plugins[capability_name] = plugin_instance
                    self.plugin_classes[capability_name] = plugin_class
                    
                    # Call on_load hook
                    plugin_instance.on_load()
                    
                    loaded_count += 1
                    print(f"  ✓ Loaded plugin: {plugin_class.__name__} → capability '{capability_name}'")
                    
                except Exception as e:
                    print(f"  ✗ Error instantiating plugin {plugin_class.__name__}: {e}")
        
        # Also load worker-to-worker plugins
        self.load_worker_to_worker_plugins()
        
        print(f"✅ Successfully loaded {loaded_count} plugins\n")
        return self.plugins
    
    def load_worker_to_worker_plugins(self):
        """Load plugins from worker-to-worker directory"""
        worker_to_worker_dir = os.path.join(os.path.dirname(__file__), 'worker-to-worker')
        
        if not os.path.exists(worker_to_worker_dir):
            return
        
        print(f"\n🔄 Loading worker-to-worker plugins from: {worker_to_worker_dir}")
        
        # Add to path if needed
        if worker_to_worker_dir not in sys.path:
            sys.path.insert(0, worker_to_worker_dir)
        
        loaded_count = 0
        for filename in os.listdir(worker_to_worker_dir):
            if filename.endswith('_plugin.py') and not filename.startswith('_'):
                module_name = filename[:-3]
                
                try:
                    # Import from worker-to-worker directory
                    import importlib.util
                    spec = importlib.util.spec_from_file_location(module_name, os.path.join(worker_to_worker_dir, filename))
                    if spec and spec.loader:
                        module = importlib.util.module_from_spec(spec)
                        spec.loader.exec_module(module)
                        
                        # Find BasePlugin classes
                        for name, obj in inspect.getmembers(module, inspect.isclass):
                            if issubclass(obj, BasePlugin) and obj is not BasePlugin:
                                plugin_instance = obj()
                                capability_name = plugin_instance.name
                                self.plugins[capability_name] = plugin_instance
                                plugin_instance.on_load()
                                loaded_count += 1
                                print(f"  ✓ Loaded: {obj.__name__} → '{capability_name}'")
                except Exception as e:
                    print(f"  ✗ Error loading {filename}: {e}")
        
        if loaded_count > 0:
            print(f"✅ Loaded {loaded_count} worker-to-worker plugins\n")
    
    def get_plugin(self, capability_name: str) -> BasePlugin:
        """Get a plugin by capability name"""
        return self.plugins.get(capability_name)
    
    def get_all_capabilities(self) -> List[dict]:
        """
        Get all capabilities as a list of capability metadata
        
        Returns:
            List of capability dictionaries for registration with Hub
        """
        capabilities = []
        
        for name, plugin in self.plugins.items():
            capability_meta = {
                "name": plugin.name,
                "description": plugin.description,
                "input_schema": plugin.input_schema,
                "output_schema": plugin.output_schema,
                "http_method": plugin.http_method,
                "accepts_file": plugin.accepts_file,
            }
            
            if plugin.accepts_file and plugin.file_field_name:
                capability_meta["file_field_name"] = plugin.file_field_name
            
            capabilities.append(capability_meta)
        
        return capabilities
    
    def execute_plugin(self, capability_name: str, params: dict, worker_sdk=None) -> dict:
        """
        Execute a plugin by capability name
        
        Args:
            capability_name: Name of the capability to execute
            params: Input parameters
            worker_sdk: Reference to worker SDK for worker-to-worker calls
            
        Returns:
            Plugin execution result
        """
        plugin = self.get_plugin(capability_name)
        
        if plugin is None:
            raise ValueError(f"Unknown capability: {capability_name}")
        
        return plugin.execute(params, worker_sdk=worker_sdk)
    
    def unload_all_plugins(self):
        """Unload all plugins and call their on_unload hooks"""
        for plugin in self.plugins.values():
            try:
                plugin.on_unload()
            except Exception as e:
                print(f"Error unloading plugin {plugin.name}: {e}")
        
        self.plugins.clear()
        self.plugin_classes.clear()
