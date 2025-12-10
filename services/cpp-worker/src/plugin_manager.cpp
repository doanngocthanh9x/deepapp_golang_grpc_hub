#include "plugin_manager.h"
#include <stdexcept>

void PluginManager::register_plugin(PluginPtr plugin) {
    plugins_[plugin->get_name()] = plugin;
}

Plugin* PluginManager::get_plugin(const std::string& name) {
    auto it = plugins_.find(name);
    if (it != plugins_.end()) {
        return it->second.get();
    }
    return nullptr;
}

std::vector<json> PluginManager::get_capabilities() {
    std::vector<json> caps;
    for (const auto& pair : plugins_) {
        json cap = {
            {"name", pair.second->get_name()},
            {"description", pair.second->get_description()},
            {"http_method", pair.second->get_http_method()},
            {"accepts_file", pair.second->accepts_file()},
            {"file_field_name", pair.second->get_file_field_name()}
        };
        caps.push_back(cap);
    }
    return caps;
}

std::vector<Plugin*> PluginManager::getAllPlugins() const {
    std::vector<Plugin*> result;
    for (const auto& pair : plugins_) {
        result.push_back(pair.second.get());
    }
    return result;
}

std::string PluginManager::execute(const std::string& capability, const std::string& params) {
    Plugin* plugin = get_plugin(capability);
    if (!plugin) {
        throw std::runtime_error("Plugin not found: " + capability);
    }
    
    json params_json = json::parse(params);
    json result = plugin->execute(params_json);
    
    return result.dump();
}
