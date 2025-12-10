#pragma once

#include "../src/plugin.h"

// Forward declarations
class HelloCppPlugin;
class StringOpsPlugin;

// Factory functions to create plugin instances
PluginPtr create_hello_cpp_plugin();
PluginPtr create_string_ops_plugin();
