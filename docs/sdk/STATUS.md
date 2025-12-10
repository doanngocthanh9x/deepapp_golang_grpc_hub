# DeepApp Golang gRPC Hub - Hướng Dẫn Sử Dụng

## Tóm Tắt

Dự án đã được tạo thành công với cấu trúc hoàn chỉnh. Server và client có thể kết nối được.

## Cách Chạy

### 1. Khởi động Server

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run cmd/hub/main.go
```

Server sẽ hiển thị:
```
✓ Server is now listening on port 50051
Server is ready to accept connections...
```

### 2. Chạy Client (trong terminal khác)

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
go run cmd/client/main.go
```

### 3. Gửi Messages

Client hỗ trợ 3 loại tin nhắn:

- **Broadcast**: `broadcast:Hello everyone!`
- **Direct**: `direct:<client_id>:Private message`
- **Channel**: `channel:news:Breaking news!`

## Trạng Thái

✅ Server khởi động thành công
✅ Client kết nối được với server  
✅ Cấu trúc project hoàn chỉnh
⚠️  Cần fix protobuf marshaling để gửi tin nhắn

## Lưu Ý

Server sẽ "treo" sau khi start - đây là hành vi bình thường vì nó đang chờ kết nối từ clients.
