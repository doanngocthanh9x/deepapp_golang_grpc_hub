cmake_minimum_required(VERSION 3.10)
project(cpp-worker)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Find packages (Ubuntu has proper config files)
find_package(Protobuf REQUIRED)
find_package(gRPC CONFIG)

# If gRPC config not found, try pkg-config
if(NOT gRPC_FOUND)
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(GRPC REQUIRED grpc++)
    pkg_check_modules(GRPC_CPP REQUIRED grpc)
endif()

# Proto files - use generated files from parent directory
# Can be overridden with -DPROTO_PATH=/custom/path
if(NOT DEFINED PROTO_PATH)
    set(PROTO_PATH "${CMAKE_CURRENT_SOURCE_DIR}/../../proto")
endif()

message(STATUS "Using proto path: ${PROTO_PATH}")

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

# Link libraries
if(gRPC_FOUND)
    target_link_libraries(cpp-worker
        gRPC::grpc++
        protobuf::libprotobuf
    )
else()
    target_link_libraries(cpp-worker
        ${GRPC_LIBRARIES}
        ${GRPC_CPP_LIBRARIES}
        ${Protobuf_LIBRARIES}
        pthread
    )
endif()
