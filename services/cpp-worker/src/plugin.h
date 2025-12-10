#pragma once

#include <string>
#include <map>
#include <vector>
#include <memory>
#include <nlohmann/json.hpp>

using json = nlohmann::json;

class ExecutionContext {
public:
    std::string worker_id;
    // Callback for calling other workers
    // std::function<json(std::string, std::string, json)> call_worker;
};

class Plugin {
public:
    virtual ~Plugin() = default;
    
    virtual std::string get_name() const = 0;
    virtual std::string get_description() const = 0;
    virtual std::string get_http_method() const { return "POST"; }
    virtual bool accepts_file() const { return false; }
    virtual std::string get_file_field_name() const { return "file"; }
    
    // New method for worker-grpc.cpp
    virtual std::string getName() const { return get_name(); }
    virtual std::string getDescription() const { return get_description(); }
    virtual std::vector<std::string> getRequiredParams() const { return {}; }
    virtual std::vector<std::string> getOptionalParams() const { return {}; }
    
    virtual json execute(const json& params, ExecutionContext* context = nullptr) = 0;
};

using PluginPtr = std::shared_ptr<Plugin>;
