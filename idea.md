Proto dùng chung (truyền JSON + file binary)
```proto
syntax = "proto3";

package ocr;

import "google/protobuf/struct.proto";

message OCRRequest {
  string type = 1;                     // "json", "file", "task"
  google.protobuf.Struct payload = 2;  // metadata, config,...
  bytes file = 3;                      // binary data
  string filename = 4;
}

message OCRResponse {
  bool ok = 1;
  google.protobuf.Struct data = 2;     // kết quả OCR dạng JSON tự do
  bytes file = 3;                      // optional: output file like mask, crop
  string message = 4;
}

service OCRService {
  rpc Run(OCRRequest) returns (OCRResponse);
}

```