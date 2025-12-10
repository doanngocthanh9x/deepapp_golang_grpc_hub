#include "../src/plugin.h"
#include <algorithm>
#include <cctype>

class StringOpsPlugin : public Plugin {
public:
    std::string get_name() const override {
        return "string_ops_cpp";
    }

    std::string get_description() const override {
        return "Perform string operations (uppercase, lowercase, reverse, length) in C++";
    }

    json execute(const json& params, ExecutionContext* context) override {
        if (!params.contains("text") || !params["text"].is_string()) {
            throw std::runtime_error("missing required parameter: text");
        }
        
        std::string text = params["text"].get<std::string>();
        std::string operation = "uppercase"; // default
        
        if (params.contains("operation") && params["operation"].is_string()) {
            operation = params["operation"].get<std::string>();
        }

        std::string result;
        
        if (operation == "uppercase") {
            result = text;
            std::transform(result.begin(), result.end(), result.begin(), ::toupper);
        } else if (operation == "lowercase") {
            result = text;
            std::transform(result.begin(), result.end(), result.begin(), ::tolower);
        } else if (operation == "reverse") {
            result = text;
            std::reverse(result.begin(), result.end());
        } else if (operation == "length") {
            return {
                {"input", text},
                {"operation", operation},
                {"result", text.length()},
                {"status", "success"}
            };
        } else {
            throw std::runtime_error("unknown operation: " + operation);
        }

        return {
            {"input", text},
            {"operation", operation},
            {"result", result},
            {"status", "success"}
        };
    }
};

// Factory function
PluginPtr create_string_ops_plugin() {
    return std::make_shared<StringOpsPlugin>();
}
