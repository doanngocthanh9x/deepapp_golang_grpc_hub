#include <iostream>
#include <string>
#include <memory>
#include <thread>
#include <atomic>
#include <chrono>
#include <csignal>
#include <grpcpp/grpcpp.h>
#include <nlohmann/json.hpp>
#include "hub.grpc.pb.h"
#include "plugin_manager.h"

// Forward declarations for plugin factory functions
extern PluginPtr create_hello_cpp_plugin();
extern PluginPtr create_string_ops_plugin();

using json = nlohmann::json;
using grpc::Channel;
using grpc::ClientContext;
using grpc::ClientReaderWriter;
using grpc::Status;

class CPPWorkerGRPC {
private:
    std::string worker_id_;
    std::string hub_address_;
    std::unique_ptr<hub::HubService::Stub> stub_;
    std::shared_ptr<ClientReaderWriter<hub::Message, hub::Message>> stream_;
    std::unique_ptr<ClientContext> context_;  // Must outlive the stream!
    std::atomic<bool> running_;
    PluginManager plugin_manager_;

    void sendRegistration() {
        std::cout << "[cpp-worker] 📝 Preparing registration message...\n";
        
        hub::Message msg;
        msg.set_type(hub::MessageType::REGISTER);
        msg.set_from(worker_id_);
        msg.set_to("hub");

        std::cout << "[cpp-worker] 📝 Building capabilities JSON...\n";
        json capabilities = json::array();
        
        // Get all plugins and create capabilities
        std::cout << "[cpp-worker] 📝 Getting plugins...\n";
        auto all_plugins = plugin_manager_.getAllPlugins();
        std::cout << "[cpp-worker] 📝 Found " << all_plugins.size() << " plugins\n";
        
        for (const auto& plugin : all_plugins) {
            std::cout << "[cpp-worker] 📝 Processing plugin: " << plugin->getName() << "\n";
            json cap = {
                {"name", plugin->getName()},
                {"description", plugin->getDescription()},
                {"http_method", "POST"},
                {"required_params", plugin->getRequiredParams()},
                {"optional_params", plugin->getOptionalParams()}
            };
            capabilities.push_back(cap);
            std::cout << "[cpp-worker] 📝 Plugin capability added\n";
        }

        std::cout << "[cpp-worker] 📝 Creating registration data JSON...\n";
        json reg_data = {
            {"worker_id", worker_id_},
            {"worker_type", "cpp"},
            {"capabilities", capabilities},
            {"status", "active"}
        };

        std::cout << "[cpp-worker] 📝 Converting JSON to string...\n";
        std::string json_str = reg_data.dump();
        std::cout << "[cpp-worker] 📝 JSON string length: " << json_str.length() << "\n";
        
        msg.set_content(json_str);
        
        std::cout << "[cpp-worker] 📤 Sending registration...\n";
        if (stream_->Write(msg)) {
            std::cout << "[cpp-worker] 📤 Sent registration with " 
                      << capabilities.size() << " capabilities\n";
        } else {
            std::cerr << "[cpp-worker] ❌ Failed to send registration\n";
        }
    }

    void handleRequest(const hub::Message& msg) {
        try {
            std::string request_id = msg.id();
            std::string original_sender = msg.from();
            
            // Parse request content
            auto content = json::parse(msg.content());
            std::string capability;
            json params;
            
            // Extract capability from metadata (protobuf Map -> JSON)
            if (!msg.metadata().empty()) {
                json metadata = json::object();
                for (const auto& pair : msg.metadata()) {
                    metadata[pair.first] = pair.second;
                }
                capability = metadata.value("capability", "");
            }
            
            // Fallback to content fields
            if (capability.empty() && content.contains("capability")) {
                capability = content["capability"];
            }
            
            if (content.contains("params")) {
                params = content["params"];
            } else {
                params = content;
            }

            std::cout << "[cpp-worker] 📨 Request: " << capability 
                      << " from " << original_sender << "\n";

            // Execute plugin
            std::string result = plugin_manager_.execute(capability, params.dump());
            
            // Send response
            sendResponse(request_id, original_sender, result);
            
        } catch (const std::exception& e) {
            std::cerr << "[cpp-worker] ❌ Error handling request: " << e.what() << "\n";
            sendError(msg.id(), msg.from(), e.what());
        }
    }

    void sendResponse(const std::string& request_id, 
                     const std::string& to, 
                     const std::string& result) {
        hub::Message response;
        response.set_type(hub::MessageType::RESPONSE);
        response.set_id(request_id);
        response.set_from(worker_id_);
        response.set_to(to);
        
        json response_data = {
            {"success", true},
            {"result", json::parse(result)}
        };
        
        response.set_content(response_data.dump());
        
        if (stream_->Write(response)) {
            std::cout << "[cpp-worker] ✅ Sent response to " << to << "\n";
        } else {
            std::cerr << "[cpp-worker] ❌ Failed to send response\n";
        }
    }

    void sendError(const std::string& request_id, 
                  const std::string& to, 
                  const std::string& error) {
        hub::Message response;
        response.set_type(hub::MessageType::RESPONSE);
        response.set_id(request_id);
        response.set_from(worker_id_);
        response.set_to(to);
        
        json error_data = {
            {"success", false},
            {"error", error}
        };
        
        response.set_content(error_data.dump());
        stream_->Write(response);
    }

    void receiveLoop() {
        hub::Message msg;
        while (running_ && stream_->Read(&msg)) {
            if (msg.type() == hub::MessageType::REQUEST) {
                handleRequest(msg);
            }
        }
    }

public:
    CPPWorkerGRPC(const std::string& worker_id, const std::string& hub_address)
        : worker_id_(worker_id), hub_address_(hub_address), running_(false) {
        
        std::cout << "[cpp-worker] 🔵 Initializing C++ Worker...\n";
        std::cout << "[cpp-worker] Worker ID: " << worker_id << "\n";
        std::cout << "[cpp-worker] Hub Address: " << hub_address << "\n";
        
        try {
            // Register plugins using factory functions
            std::cout << "[cpp-worker] Registering hello_cpp plugin...\n";
            plugin_manager_.register_plugin(create_hello_cpp_plugin());
            
            std::cout << "[cpp-worker] Registering string_ops plugin...\n";
            plugin_manager_.register_plugin(create_string_ops_plugin());
            
            std::cout << "[cpp-worker] ✅ Plugins registered successfully\n";
        } catch (const std::exception& e) {
            std::cerr << "[cpp-worker] ❌ Error in constructor: " << e.what() << "\n";
            throw;
        }
    }

    bool connect() {
        try {
            std::cout << "[cpp-worker] Connecting to Hub at " << hub_address_ << "...\n";
            
            auto channel = grpc::CreateChannel(hub_address_, 
                                              grpc::InsecureChannelCredentials());
            
            if (!channel) {
                std::cerr << "[cpp-worker] ❌ Failed to create gRPC channel\n";
                return false;
            }
            
            std::cout << "[cpp-worker] ✓ Channel created\n";
            
            stub_ = hub::HubService::NewStub(channel);
            
            if (!stub_) {
                std::cerr << "[cpp-worker] ❌ Failed to create stub\n";
                return false;
            }
            
            std::cout << "[cpp-worker] ✓ Stub created\n";
            
            // Context must outlive the stream!
            context_ = std::make_unique<ClientContext>();
            stream_ = stub_->Connect(context_.get());
            
            if (!stream_) {
                std::cerr << "[cpp-worker] ❌ Failed to connect stream\n";
                return false;
            }
            
            std::cout << "[cpp-worker] ✓ Connected to Hub\n";
            return true;
            
        } catch (const std::exception& e) {
            std::cerr << "[cpp-worker] ❌ Exception in connect(): " << e.what() << "\n";
            return false;
        }
    }

    void run() {
        running_ = true;
        
        // Send registration
        sendRegistration();
        
        std::cout << "[cpp-worker] 📨 Listening for requests...\n";
        
        // Start receive loop
        receiveLoop();
        
        running_ = false;
        stream_->WritesDone();
        Status status = stream_->Finish();
        
        if (!status.ok()) {
            std::cerr << "[cpp-worker] Connection error: " 
                      << status.error_message() << "\n";
        }
    }

    void shutdown() {
        running_ = false;
    }
};

// Signal handling
std::unique_ptr<CPPWorkerGRPC> worker_instance;

void signalHandler(int signum) {
    std::cout << "\n[cpp-worker] Received signal " << signum << ", shutting down...\n";
    if (worker_instance) {
        worker_instance->shutdown();
    }
    exit(signum);
}

int main() {
    // Explicitly set unbuffered output for debugging
    std::cout.setf(std::ios::unitbuf);
    std::cerr.setf(std::ios::unitbuf);
    
    std::cout << "[cpp-worker] 🚀 Starting C++ Worker (step 1)...\n";
    
    // Initialize Google's logging library used by gRPC/protobuf
    GOOGLE_PROTOBUF_VERIFY_VERSION;
    
    try {
        std::cout << "[cpp-worker] 🚀 Step 2: Setting up signal handlers...\n";
        
        signal(SIGINT, signalHandler);
        signal(SIGTERM, signalHandler);
        
        std::cout << "[cpp-worker] 🚀 Step 3: Declaring variables...\n" << std::flush;
        const std::string worker_id = "cpp-worker";
        const std::string hub_address = "localhost:50051";
        
        std::cout << "[cpp-worker] 🚀 Step 4: Creating worker instance...\n" << std::flush;
        worker_instance = std::make_unique<CPPWorkerGRPC>(worker_id, hub_address);
        std::cout << "[cpp-worker] ✅ Worker instance created\n" << std::flush;
        
        // Retry connection
        int max_retries = 10;
        int retry_delay = 2;
        
        for (int i = 0; i < max_retries; i++) {
            if (worker_instance->connect()) {
                worker_instance->run();
                break;
            }
            
            std::cout << "[cpp-worker] Retry " << (i+1) << "/" << max_retries 
                      << " in " << retry_delay << "s...\n";
            std::this_thread::sleep_for(std::chrono::seconds(retry_delay));
        }
        
        std::cout << "[cpp-worker] Worker finished\n";
        return 0;
        
    } catch (const std::exception& e) {
        std::cerr << "[cpp-worker] ❌ Fatal error: " << e.what() << "\n";
        return 1;
    } catch (...) {
        std::cerr << "[cpp-worker] ❌ Unknown fatal error\n";
        return 1;
    }
}
