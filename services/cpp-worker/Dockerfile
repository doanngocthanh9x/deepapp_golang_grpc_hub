# Dockerfile for C++ Worker with Ubuntu (better gRPC support than Alpine)
FROM ubuntu:22.04 AS builder

# Avoid interactive prompts
ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    pkg-config \
    libgrpc++-dev \
    libprotobuf-dev \
    protobuf-compiler-grpc \
    nlohmann-json3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy proto files (from parent context)
COPY proto ./proto

# Generate C++ proto files
RUN cd proto && \
    protoc --cpp_out=. --grpc_out=. \
    --plugin=protoc-gen-grpc=/usr/bin/grpc_cpp_plugin \
    -I. hub.proto && \
    ls -la *.pb.* || echo "Proto files generated"

# Copy C++ worker source
COPY services/cpp-worker ./services/cpp-worker

# Build with CMake - fix path in CMakeLists to use generated files
RUN cd services/cpp-worker && \
    mkdir -p build && \
    cd build && \
    cmake -DPROTO_PATH=/app/proto .. && \
    make -j$(nproc)

# Runtime stage - minimal Ubuntu
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    libgrpc++1.30 \
    libprotobuf23 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/services/cpp-worker/build/cpp-worker /usr/local/bin/cpp-worker

ENTRYPOINT ["/usr/local/bin/cpp-worker"]
