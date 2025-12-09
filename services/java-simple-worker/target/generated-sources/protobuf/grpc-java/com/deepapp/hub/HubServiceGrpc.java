package com.deepapp.hub;

import static io.grpc.MethodDescriptor.generateFullMethodName;

/**
 */
@javax.annotation.Generated(
    value = "by gRPC proto compiler (version 1.60.0)",
    comments = "Source: hub.proto")
@io.grpc.stub.annotations.GrpcGenerated
public final class HubServiceGrpc {

  private HubServiceGrpc() {}

  public static final java.lang.String SERVICE_NAME = "hub.HubService";

  // Static method descriptors that strictly reflect the proto.
  private static volatile io.grpc.MethodDescriptor<com.deepapp.hub.Hub.Message,
      com.deepapp.hub.Hub.Message> getConnectMethod;

  @io.grpc.stub.annotations.RpcMethod(
      fullMethodName = SERVICE_NAME + '/' + "Connect",
      requestType = com.deepapp.hub.Hub.Message.class,
      responseType = com.deepapp.hub.Hub.Message.class,
      methodType = io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
  public static io.grpc.MethodDescriptor<com.deepapp.hub.Hub.Message,
      com.deepapp.hub.Hub.Message> getConnectMethod() {
    io.grpc.MethodDescriptor<com.deepapp.hub.Hub.Message, com.deepapp.hub.Hub.Message> getConnectMethod;
    if ((getConnectMethod = HubServiceGrpc.getConnectMethod) == null) {
      synchronized (HubServiceGrpc.class) {
        if ((getConnectMethod = HubServiceGrpc.getConnectMethod) == null) {
          HubServiceGrpc.getConnectMethod = getConnectMethod =
              io.grpc.MethodDescriptor.<com.deepapp.hub.Hub.Message, com.deepapp.hub.Hub.Message>newBuilder()
              .setType(io.grpc.MethodDescriptor.MethodType.BIDI_STREAMING)
              .setFullMethodName(generateFullMethodName(SERVICE_NAME, "Connect"))
              .setSampledToLocalTracing(true)
              .setRequestMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.deepapp.hub.Hub.Message.getDefaultInstance()))
              .setResponseMarshaller(io.grpc.protobuf.ProtoUtils.marshaller(
                  com.deepapp.hub.Hub.Message.getDefaultInstance()))
              .setSchemaDescriptor(new HubServiceMethodDescriptorSupplier("Connect"))
              .build();
        }
      }
    }
    return getConnectMethod;
  }

  /**
   * Creates a new async stub that supports all call types for the service
   */
  public static HubServiceStub newStub(io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<HubServiceStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<HubServiceStub>() {
        @java.lang.Override
        public HubServiceStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new HubServiceStub(channel, callOptions);
        }
      };
    return HubServiceStub.newStub(factory, channel);
  }

  /**
   * Creates a new blocking-style stub that supports unary and streaming output calls on the service
   */
  public static HubServiceBlockingStub newBlockingStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<HubServiceBlockingStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<HubServiceBlockingStub>() {
        @java.lang.Override
        public HubServiceBlockingStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new HubServiceBlockingStub(channel, callOptions);
        }
      };
    return HubServiceBlockingStub.newStub(factory, channel);
  }

  /**
   * Creates a new ListenableFuture-style stub that supports unary calls on the service
   */
  public static HubServiceFutureStub newFutureStub(
      io.grpc.Channel channel) {
    io.grpc.stub.AbstractStub.StubFactory<HubServiceFutureStub> factory =
      new io.grpc.stub.AbstractStub.StubFactory<HubServiceFutureStub>() {
        @java.lang.Override
        public HubServiceFutureStub newStub(io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
          return new HubServiceFutureStub(channel, callOptions);
        }
      };
    return HubServiceFutureStub.newStub(factory, channel);
  }

  /**
   */
  public interface AsyncService {

    /**
     */
    default io.grpc.stub.StreamObserver<com.deepapp.hub.Hub.Message> connect(
        io.grpc.stub.StreamObserver<com.deepapp.hub.Hub.Message> responseObserver) {
      return io.grpc.stub.ServerCalls.asyncUnimplementedStreamingCall(getConnectMethod(), responseObserver);
    }
  }

  /**
   * Base class for the server implementation of the service HubService.
   */
  public static abstract class HubServiceImplBase
      implements io.grpc.BindableService, AsyncService {

    @java.lang.Override public final io.grpc.ServerServiceDefinition bindService() {
      return HubServiceGrpc.bindService(this);
    }
  }

  /**
   * A stub to allow clients to do asynchronous rpc calls to service HubService.
   */
  public static final class HubServiceStub
      extends io.grpc.stub.AbstractAsyncStub<HubServiceStub> {
    private HubServiceStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected HubServiceStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new HubServiceStub(channel, callOptions);
    }

    /**
     */
    public io.grpc.stub.StreamObserver<com.deepapp.hub.Hub.Message> connect(
        io.grpc.stub.StreamObserver<com.deepapp.hub.Hub.Message> responseObserver) {
      return io.grpc.stub.ClientCalls.asyncBidiStreamingCall(
          getChannel().newCall(getConnectMethod(), getCallOptions()), responseObserver);
    }
  }

  /**
   * A stub to allow clients to do synchronous rpc calls to service HubService.
   */
  public static final class HubServiceBlockingStub
      extends io.grpc.stub.AbstractBlockingStub<HubServiceBlockingStub> {
    private HubServiceBlockingStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected HubServiceBlockingStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new HubServiceBlockingStub(channel, callOptions);
    }
  }

  /**
   * A stub to allow clients to do ListenableFuture-style rpc calls to service HubService.
   */
  public static final class HubServiceFutureStub
      extends io.grpc.stub.AbstractFutureStub<HubServiceFutureStub> {
    private HubServiceFutureStub(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      super(channel, callOptions);
    }

    @java.lang.Override
    protected HubServiceFutureStub build(
        io.grpc.Channel channel, io.grpc.CallOptions callOptions) {
      return new HubServiceFutureStub(channel, callOptions);
    }
  }

  private static final int METHODID_CONNECT = 0;

  private static final class MethodHandlers<Req, Resp> implements
      io.grpc.stub.ServerCalls.UnaryMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ServerStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.ClientStreamingMethod<Req, Resp>,
      io.grpc.stub.ServerCalls.BidiStreamingMethod<Req, Resp> {
    private final AsyncService serviceImpl;
    private final int methodId;

    MethodHandlers(AsyncService serviceImpl, int methodId) {
      this.serviceImpl = serviceImpl;
      this.methodId = methodId;
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public void invoke(Req request, io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        default:
          throw new AssertionError();
      }
    }

    @java.lang.Override
    @java.lang.SuppressWarnings("unchecked")
    public io.grpc.stub.StreamObserver<Req> invoke(
        io.grpc.stub.StreamObserver<Resp> responseObserver) {
      switch (methodId) {
        case METHODID_CONNECT:
          return (io.grpc.stub.StreamObserver<Req>) serviceImpl.connect(
              (io.grpc.stub.StreamObserver<com.deepapp.hub.Hub.Message>) responseObserver);
        default:
          throw new AssertionError();
      }
    }
  }

  public static final io.grpc.ServerServiceDefinition bindService(AsyncService service) {
    return io.grpc.ServerServiceDefinition.builder(getServiceDescriptor())
        .addMethod(
          getConnectMethod(),
          io.grpc.stub.ServerCalls.asyncBidiStreamingCall(
            new MethodHandlers<
              com.deepapp.hub.Hub.Message,
              com.deepapp.hub.Hub.Message>(
                service, METHODID_CONNECT)))
        .build();
  }

  private static abstract class HubServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoFileDescriptorSupplier, io.grpc.protobuf.ProtoServiceDescriptorSupplier {
    HubServiceBaseDescriptorSupplier() {}

    @java.lang.Override
    public com.google.protobuf.Descriptors.FileDescriptor getFileDescriptor() {
      return com.deepapp.hub.Hub.getDescriptor();
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.ServiceDescriptor getServiceDescriptor() {
      return getFileDescriptor().findServiceByName("HubService");
    }
  }

  private static final class HubServiceFileDescriptorSupplier
      extends HubServiceBaseDescriptorSupplier {
    HubServiceFileDescriptorSupplier() {}
  }

  private static final class HubServiceMethodDescriptorSupplier
      extends HubServiceBaseDescriptorSupplier
      implements io.grpc.protobuf.ProtoMethodDescriptorSupplier {
    private final java.lang.String methodName;

    HubServiceMethodDescriptorSupplier(java.lang.String methodName) {
      this.methodName = methodName;
    }

    @java.lang.Override
    public com.google.protobuf.Descriptors.MethodDescriptor getMethodDescriptor() {
      return getServiceDescriptor().findMethodByName(methodName);
    }
  }

  private static volatile io.grpc.ServiceDescriptor serviceDescriptor;

  public static io.grpc.ServiceDescriptor getServiceDescriptor() {
    io.grpc.ServiceDescriptor result = serviceDescriptor;
    if (result == null) {
      synchronized (HubServiceGrpc.class) {
        result = serviceDescriptor;
        if (result == null) {
          serviceDescriptor = result = io.grpc.ServiceDescriptor.newBuilder(SERVICE_NAME)
              .setSchemaDescriptor(new HubServiceFileDescriptorSupplier())
              .addMethod(getConnectMethod())
              .build();
        }
      }
    }
    return result;
  }
}
