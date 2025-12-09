"""
Plugin loader - Tự động load capabilities từ plugins/
"""

import os
import sys
import importlib.util
from pathlib import Path
from typing import Dict, Any
from decorators import get_registered_capabilities


class PluginLoader:
    """Load plugins from a directory"""
    
    def __init__(self, plugins_dir: str = "plugins"):
        self.plugins_dir = plugins_dir
        self.loaded_plugins = []
        
    def load_plugins(self) -> Dict[str, Dict[str, Any]]:
        """
        Load all Python files from plugins directory
        Returns dict of capabilities with their handlers
        """
        plugins_path = Path(self.plugins_dir)
        
        if not plugins_path.exists():
            print(f"⚠️  Plugins directory not found: {self.plugins_dir}")
            return {}
        
        # Add plugins dir to Python path
        sys.path.insert(0, str(plugins_path.absolute()))
        
        # Load all .py files
        plugin_files = list(plugins_path.glob("*.py"))
        
        if not plugin_files:
            print(f"⚠️  No plugin files found in {self.plugins_dir}")
            return {}
        
        print(f"🔍 Loading plugins from {self.plugins_dir}...")
        
        for plugin_file in plugin_files:
            if plugin_file.name.startswith("_"):
                continue  # Skip __init__.py and private files
                
            try:
                self._load_plugin_file(plugin_file)
            except Exception as e:
                print(f"❌ Failed to load plugin {plugin_file.name}: {e}")
        
        capabilities = get_registered_capabilities()
        print(f"✅ Loaded {len(capabilities)} capabilities from {len(self.loaded_plugins)} plugins")
        
        return capabilities
    
    def _load_plugin_file(self, plugin_file: Path):
        """Load a single plugin file"""
        module_name = plugin_file.stem
        
        # Import module
        spec = importlib.util.spec_from_file_location(module_name, plugin_file)
        if spec is None or spec.loader is None:
            raise ImportError(f"Cannot load spec for {plugin_file}")
        
        module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(module)
        
        self.loaded_plugins.append(module_name)
        print(f"  📦 Loaded plugin: {module_name}")
