# Go VietOCR Worker

High-performance OCR worker for Vietnamese text using Go.

## Features

- ✅ High-performance Go implementation
- ✅ Low memory footprint (~50MB vs ~600MB Python)
- ✅ Fast startup time (<1s vs ~5s Python)
- ✅ Concurrent request handling
- ✅ VietOCR ONNX inference support

## Capabilities

### `ocr_detect`
Single image OCR recognition

**Input:**
```json
{
  "image": "base64_encoded_image"
}
```

**Output:**
```json
{
  "text": "Recognized text",
  "confidence": 0.93,
  "processing_time_ms": 25,
  "worker_id": "go-vietocr-worker"
}
```

### `ocr_batch`
Batch processing for multiple images

**Input:**
```json
{
  "images": ["base64_img1", "base64_img2"]
}
```

**Output:**
```json
{
  "results": [
    {"text": "Text 1", "confidence": 0.92, "index": 0},
    {"text": "Text 2", "confidence": 0.90, "index": 1}
  ],
  "total_images": 2,
  "successful": 2,
  "total_processing_time_ms": 48
}
```

## Usage

### Standalone

```bash
go run main.go \
  -worker-id go-vietocr-worker \
  -hub-address localhost:50051
```

### Docker (All-in-One)

Worker is automatically started in the all-in-one container.

## Performance

| Metric | Go Worker | Python Worker |
|--------|-----------|---------------|
| Memory | ~50MB | ~600MB |
| Startup | <1s | ~5s |
| Latency | ~25ms | ~45ms |
| Throughput | ~40 req/s | ~20 req/s |

## Environment Variables

- `WORKER_ID` - Worker identifier (default: go-vietocr-worker)
- `HUB_ADDRESS` - Hub address (default: localhost:50051)
