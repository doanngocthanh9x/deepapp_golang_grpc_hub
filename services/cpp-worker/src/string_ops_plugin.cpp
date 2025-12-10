#include "plugin.h"
#include <algorithm>
#include <cctype>
#include <string>

class StringOpsPlugin : public Plugin {
public:
    std::string get_name() const override {
        return "string_ops";
    }

    std::string get_description() const override {
        return "String manipulation operations (upper, lower, reverse)";
    }

    json execute(const json& input, ExecutionContext* context = nullptr) override {
        std::string text = input.value("text", "");
        std::string operation = input.value("operation", "");
        std::string result = text;

        if (operation == "upper") {
            std::transform(result.begin(), result.end(), result.begin(),
                         [](unsigned char c) { return std::toupper(c); });
        } else if (operation == "lower") {
            std::transform(result.begin(), result.end(), result.begin(),
                         [](unsigned char c) { return std::tolower(c); });
        } else if (operation == "reverse") {
            std::reverse(result.begin(), result.end());
        } else {
            throw std::runtime_error("Unknown operation: " + operation);
        }

        return {{"result", result}};
    }
};

PluginPtr create_string_ops_plugin() {
    return std::make_shared<StringOpsPlugin>();
}
