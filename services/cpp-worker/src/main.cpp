#include "worker.h"
#include <iostream>
#include <csignal>
#include <memory>

// Forward declarations of plugin factory functions
extern "C" PluginPtr create_hello_cpp_plugin();
extern "C" PluginPtr create_string_ops_plugin();

std::unique_ptr<CPPWorker> g_worker;

void signal_handler(int signal) {
    if (g_worker) {
        g_worker->shutdown();
    }
}

int main(int argc, char** argv) {
    try {
        // Setup signal handlers
        std::signal(SIGINT, signal_handler);
        std::signal(SIGTERM, signal_handler);

        // Get hub address from environment or use default
        const char* hub_addr_env = std::getenv("HUB_ADDRESS");
        std::string hub_address = hub_addr_env ? hub_addr_env : "localhost:50051";

        std::cout << "[cpp-worker] 🔵 C++ Worker Starting..." << std::endl;
        std::cout << "[cpp-worker] Hub address: " << hub_address << std::endl;

        // Create worker
        g_worker = std::make_unique<CPPWorker>("cpp-worker", hub_address);

        // Register plugins
        std::cout << "[cpp-worker] Registering plugins..." << std::endl;
        g_worker->register_plugin(create_hello_cpp_plugin());
        g_worker->register_plugin(create_string_ops_plugin());

        // Connect to hub
        std::cout << "[cpp-worker] Connecting to Hub..." << std::endl;
        if (!g_worker->connect()) {
            std::cerr << "[cpp-worker] ❌ Failed to connect to Hub" << std::endl;
            return 1;
        }

        // Run worker
        std::cout << "[cpp-worker] ✅ C++ Worker ready!" << std::endl;
        g_worker->run();

    } catch (const std::exception& e) {
        std::cerr << "[cpp-worker] ❌ Exception: " << e.what() << std::endl;
        return 1;
    } catch (...) {
        std::cerr << "[cpp-worker] ❌ Unknown exception" << std::endl;
        return 1;
    }

    return 0;
}
