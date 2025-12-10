#pragma once

#include "plugin.h"
#include <vector>
#include <unordered_map>
#include <memory>

class PluginManager {
public:
    void register_plugin(PluginPtr plugin);
    Plugin* get_plugin(const std::string& name);
    std::vector<json> get_capabilities();
    std::vector<Plugin*> getAllPlugins() const;
    std::string execute(const std::string& capability, const std::string& params);
    size_t plugin_count() const { return plugins_.size(); }

private:
    std::unordered_map<std::string, PluginPtr> plugins_;
};
