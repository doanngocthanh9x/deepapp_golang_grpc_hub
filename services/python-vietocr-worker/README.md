# Python VietOCR Worker

Worker OCR cho tiếng Việt sử dụng ONNX Runtime.

## 📋 Tổng quan

Worker này tích hợp VietOCR với ONNX Runtime để nhận diện văn bản tiếng Việt từ ảnh. Tối ưu hiệu năng và dễ dàng triển khai.

## 🎯 Capabilities

### `ocr_detect` - OCR đơn lẻ
Nhận diện text từ một ảnh

**Input:**
```json
{
  "image": "base64_encoded_image"
}
```

**Output:**
```json
{
  "text": "Nhận diện văn bản tiếng Việt",
  "confidence": 0.95,
  "processing_time_ms": 45.2,
  "status": "success"
}
```

### `ocr_batch` - OCR batch
Xử lý nhiều ảnh cùng lúc

**Input:**
```json
{
  "images": ["base64_img1", "base64_img2", "base64_img3"]
}
```

**Output:**
```json
{
  "results": [
    {"text": "Text 1", "confidence": 0.95, "index": 0},
    {"text": "Text 2", "confidence": 0.92, "index": 1}
  ],
  "total_images": 2,
  "successful": 2,
  "total_processing_time_ms": 120.5,
  "status": "success"
}
```

## 🚀 Sử dụng

### Với All-in-One Container

Worker này đã được tích hợp sẵn trong container all-in-one:

```bash
cd /home/vps1/WorkSpace/deepapp_golang_grpc_hub
docker-compose -f docker-compose.all-in-one.yml up
```

### Chạy độc lập

```bash
cd services/python-vietocr-worker

# Install dependencies
pip install -r requirements.txt

# Run worker
python vietocr_worker.py
```

### Environment Variables

```bash
WORKER_ID=python-vietocr-worker
HUB_ADDRESS=localhost:50051
ENCODER_PATH=/app/models/transformer_encoder.onnx
DECODER_PATH=/app/models/transformer_decoder.onnx
USE_GPU=false
```

## 📦 Convert VietOCR Models

Để sử dụng model thực tế (không phải demo mode), cần convert VietOCR model sang ONNX:

### 1. Install conversion tools

```bash
pip install torch vietocr onnx onnx-simplifier
```

### 2. Download VietOCR checkpoint

Tải pretrained model từ [VietOCR repo](https://github.com/pbcquoc/vietocr)

### 3. Convert to ONNX

Sử dụng script conversion từ [vietocr-tensorrt](https://github.com/NNDam/vietocr-tensorrt):

```bash
# Clone vietocr-tensorrt
git clone https://github.com/NNDam/vietocr-tensorrt.git
cd vietocr-tensorrt

# Convert model
python convert.py \
    --checkpoint path/to/vietocr_checkpoint.pth \
    --output-dir ./onnx_models \
    --simplify
```

Output models:
- `transformer_encoder.onnx` - CNN + Transformer Encoder
- `transformer_decoder.onnx` - Transformer Decoder

### 4. Mount models vào container

Sửa `docker-compose.all-in-one.yml`:

```yaml
services:
  deepapp-hub:
    volumes:
      - hub-data:/data
      - ./models:/app/models:ro  # Mount ONNX models
```

Hoặc copy models vào container:

```bash
docker cp transformer_encoder.onnx deepapp-hub-all-in-one:/app/models/
docker cp transformer_decoder.onnx deepapp-hub-all-in-one:/app/models/
docker restart deepapp-hub-all-in-one
```

## 📡 API Usage

### Via Web API

```bash
# Encode image to base64
IMAGE_BASE64=$(base64 -w 0 test_image.jpg)

# Call OCR
curl -X POST http://localhost:8081/api/call \
  -H 'Content-Type: application/json' \
  -d "{
    \"worker_id\": \"python-vietocr-worker\",
    \"capability\": \"ocr_detect\",
    \"data\": {
      \"image\": \"$IMAGE_BASE64\"
    }
  }"
```

### Response

```json
{
  "success": true,
  "data": {
    "text": "Nhận diện văn bản tiếng Việt",
    "confidence": 0.95,
    "processing_time_ms": 45.2,
    "worker_id": "python-vietocr-worker",
    "status": "success"
  }
}
```

## ⚡ Performance

- **Latency**: ~50ms (CPU), ~25ms (GPU)
- **Memory**: ~500-600MB
- **Throughput**: ~20 FPS (CPU), ~40 FPS (GPU)

## 🔧 Demo Mode

Worker chạy ở **demo mode** nếu ONNX models không được tìm thấy:
- Trả về text demo thay vì nhận diện thực
- Confidence cố định 0.95
- Vẫn có thể test workflow và API

## 🐛 Troubleshooting

### Models not found

```bash
# Check logs
docker logs deepapp-hub-all-in-one | grep vietocr

# Should see:
# ⚠️  VietOCR models not found - running in demo mode
```

**Solution**: Convert và mount ONNX models (xem phần Convert Models)

### CUDA not available

```bash
# Install ONNX Runtime GPU
pip uninstall onnxruntime
pip install onnxruntime-gpu

# Enable GPU in environment
USE_GPU=true
```

### Connection refused

```bash
# Check Hub is running
docker ps | grep deepapp-hub

# Check worker logs
docker logs deepapp-hub-all-in-one | grep python-vietocr
```

## 📚 References

- [VietOCR](https://github.com/pbcquoc/vietocr) - Original VietOCR implementation
- [vietocr-tensorrt](https://github.com/NNDam/vietocr-tensorrt) - TensorRT/ONNX conversion
- [ONNX Runtime](https://onnxruntime.ai/) - Inference engine
- [DeepApp Hub](../../README.md) - gRPC Hub documentation

## 📄 License

MIT License
