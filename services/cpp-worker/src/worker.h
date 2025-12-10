#pragma once

#include "plugin_manager.h"
#include <grpcpp/grpcpp.h>
#include <string>
#include <memory>
#include <atomic>

// Forward declaration
namespace hub { class Message; }

class CPPWorker {
public:
    CPPWorker(const std::string& worker_id, const std::string& hub_address);
    ~CPPWorker();

    void register_plugin(PluginPtr plugin);
    bool connect();
    void run();
    void shutdown();

private:
    void send_registration();
    void receive_messages();
    void handle_message(const hub::Message& msg);
    void handle_request(const hub::Message& msg);
    void send_response(const std::string& request_id, 
                      const std::string& target_client,
                      const json& result);
    void send_error_response(const std::string& request_id,
                            const std::string& target_client,
                            const std::string& error_msg);
    void start_heartbeat();

    std::string worker_id_;
    std::string hub_address_;
    PluginManager plugin_manager_;
    
    std::shared_ptr<grpc::Channel> channel_;
    // std::unique_ptr<hub::HubService::Stub> stub_;
    // std::shared_ptr<grpc::ClientReaderWriter<hub::Message, hub::Message>> stream_;
    
    std::atomic<bool> connected_{false};
    std::atomic<bool> running_{false};
};
