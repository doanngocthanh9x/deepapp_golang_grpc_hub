#include "plugin.h"
#include <string>

class HelloCppPlugin : public Plugin {
public:
    std::string get_name() const override {
        return "hello_cpp";
    }

    std::string get_description() const override {
        return "Returns a friendly hello message";
    }

    json execute(const json& input, ExecutionContext* context = nullptr) override {
        std::string name = input.value("name", "World");
        return {
            {"message", "Hello from C++, " + name + "!"}
        };
    }
};

PluginPtr create_hello_cpp_plugin() {
    return std::make_shared<HelloCppPlugin>();
}
