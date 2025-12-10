#include <iostream>
#include <grpcpp/grpcpp.h>

int main() {
    std::cout << "Starting minimal C++ worker test...\n";
    
    try {
        std::cout << "Creating channel...\n";
        auto channel = grpc::CreateChannel("localhost:50051", 
                                          grpc::InsecureChannelCredentials());
        std::cout << "Channel created successfully!\n";
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << "\n";
        return 1;
    }
}
