#include "../src/plugin.h"
#include <chrono>
#include <iomanip>
#include <sstream>

class HelloCppPlugin : public Plugin {
public:
    std::string get_name() const override {
        return "hello_cpp";
    }

    std::string get_description() const override {
        return "Returns a hello message from C++ worker";
    }

    json execute(const json& params, ExecutionContext* context) override {
        std::string name = "World";
        if (params.contains("name") && params["name"].is_string()) {
            name = params["name"].get<std::string>();
        }

        // Get current timestamp
        auto now = std::chrono::system_clock::now();
        auto time_t_now = std::chrono::system_clock::to_time_t(now);
        std::stringstream ss;
        ss << std::put_time(std::gmtime(&time_t_now), "%Y-%m-%dT%H:%M:%SZ");

        return {
            {"message", "Hello " + name + " from C++! 🔷"},
            {"worker_id", context->worker_id},
            {"timestamp", ss.str()},
            {"cpp_version", __cplusplus},
            {"status", "success"}
        };
    }
};

// Factory function
PluginPtr create_hello_cpp_plugin() {
    return std::make_shared<HelloCppPlugin>();
}
