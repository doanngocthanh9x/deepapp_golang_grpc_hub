cmake_minimum_required(VERSION 3.10)
project(cpp-worker)

# Use C++17 standard as per Ubuntu guide
set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -Wextra -g")
set(CMAKE_CXX_FLAGS_DEBUG "${CMAKE_CXX_FLAGS_DEBUG} -O0 -g")
set(CMAKE_CXX_FLAGS_RELEASE "${CMAKE_CXX_FLAGS_RELEASE} -O2")

# Find packages (Ubuntu has proper config files)
find_package(Protobuf REQUIRED)
find_package(gRPC CONFIG)

# If gRPC config not found, try pkg-config
if(NOT gRPC_FOUND)
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(GRPC REQUIRED grpc++)
    pkg_check_modules(GRPC_CPP REQUIRED grpc++)
    pkg_check_modules(PROTOBUF REQUIRED protobuf)
endif()

# Proto files - use generated files from parent directory
# Can be overridden with -DPROTO_PATH=/custom/path
if(NOT DEFINED PROTO_PATH)
    set(PROTO_PATH "${CMAKE_CURRENT_SOURCE_DIR}/../../proto")
endif()

message(STATUS "Using proto path: ${PROTO_PATH}")
message(STATUS "Using C++ standard: C++${CMAKE_CXX_STANDARD}")

# Include directories
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/src)
include_directories(${PROTO_PATH})
include_directories(${Protobuf_INCLUDE_DIRS})
if(GRPC_INCLUDE_DIRS)
    include_directories(${GRPC_INCLUDE_DIRS})
endif()

# Source files
set(SOURCES
    src/worker-grpc.cpp
    src/plugin_manager.cpp
    src/hello_plugin.cpp
    src/string_ops_plugin.cpp
)

# Proto generated sources
set(PROTO_SRCS
    ${PROTO_PATH}/hub.pb.cc
    ${PROTO_PATH}/hub.grpc.pb.cc
)

# Create executable
add_executable(cpp-worker ${SOURCES} ${PROTO_SRCS})

# Create minimal test executable
add_executable(cpp-worker-minimal src/main-minimal.cpp)

# Link libraries with proper dependencies
if(gRPC_FOUND)
    target_link_libraries(cpp-worker
        gRPC::grpc++
        protobuf::libprotobuf
        pthread
    )
    target_link_libraries(cpp-worker-minimal
        gRPC::grpc++
        pthread
    )
else()
    target_link_libraries(cpp-worker
        ${GRPC_LIBRARIES}
        ${PROTOBUF_LIBRARIES}
        pthread
        dl
    )
    target_link_libraries(cpp-worker-minimal
        ${GRPC_LIBRARIES}
        pthread
        dl
    )
endif()

# Add installation rules
install(TARGETS cpp-worker DESTINATION bin)
install(TARGETS cpp-worker-minimal DESTINATION bin)
