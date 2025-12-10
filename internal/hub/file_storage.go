package hub

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"deepapp_golang_grpc_hub/internal/proto"
)

// FileStorage handles file upload/download with chunking
type FileStorage struct {
	mu        sync.RWMutex
	storePath string
	files     map[string]*FileInfo // file_id -> FileInfo
}

// FileInfo stores metadata about uploaded files
type FileInfo struct {
	FileID    string
	Filename  string
	Size      int64
	MimeType  string
	Path      string
	CreatedAt string
}

// NewFileStorage creates a new file storage
func NewFileStorage(storePath string) *FileStorage {
	// Create storage directory if not exists
	os.MkdirAll(storePath, 0755)
	
	return &FileStorage{
		storePath: storePath,
		files:     make(map[string]*FileInfo),
	}
}

// UploadFile handles streaming file upload
func (s *Server) UploadFile(stream proto.HubService_UploadFileServer) error {
	var fileID string
	var filename string
	var filePath string
	var file *os.File
	var totalReceived int64
	var err error

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// Upload complete
			break
		}
		if err != nil {
			return fmt.Errorf("failed to receive chunk: %v", err)
		}

		// First chunk - create file
		if file == nil {
			fileID = chunk.FileId
			filename = chunk.Metadata["filename"]
			if filename == "" {
				filename = fileID
			}

			filePath = filepath.Join("/tmp/hub_files", fileID)
			os.MkdirAll(filepath.Dir(filePath), 0755)

			file, err = os.Create(filePath)
			if err != nil {
				return fmt.Errorf("failed to create file: %v", err)
			}
			defer file.Close()

			fmt.Printf("📥 Receiving file: %s (%d bytes)\n", filename, chunk.TotalSize)
		}

		// Write chunk
		n, err := file.Write(chunk.Data)
		if err != nil {
			return fmt.Errorf("failed to write chunk: %v", err)
		}

		totalReceived += int64(n)

		fmt.Printf("📦 Received chunk: %d/%d bytes (%.1f%%)\n",
			totalReceived, chunk.TotalSize,
			float64(totalReceived)/float64(chunk.TotalSize)*100)
	}

	if file != nil {
		file.Close()
		fmt.Printf("✅ File upload complete: %s (%d bytes)\n", filename, totalReceived)
	}

	// Send response
	return stream.SendAndClose(&proto.FileUploadResponse{
		FileId:  fileID,
		Success: true,
		Message: "File uploaded successfully",
		Size:    totalReceived,
	})
}

// DownloadFile handles streaming file download
func (s *Server) DownloadFile(req *proto.FileDownloadRequest, stream proto.HubService_DownloadFileServer) error {
	fileID := req.FileId
	filePath := filepath.Join("/tmp/hub_files", fileID)

	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file not found: %s", fileID)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fmt.Printf("📤 Sending file: %s (%d bytes)\n", fileID, fileInfo.Size())

	// Determine chunk size
	chunkSize := req.ChunkSize
	if chunkSize == 0 {
		chunkSize = 64 * 1024 // 64KB default
	}

	// Seek to offset if specified
	if req.Offset > 0 {
		_, err = file.Seek(req.Offset, 0)
		if err != nil {
			return fmt.Errorf("failed to seek: %v", err)
		}
	}

	// Stream file in chunks
	buffer := make([]byte, chunkSize)
	var offset int64 = req.Offset
	var sentChunks int

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}

		chunk := &proto.FileChunk{
			FileId:    fileID,
			Data:      buffer[:n],
			Offset:    offset,
			TotalSize: fileInfo.Size(),
			IsLast:    false,
		}

		if err := stream.Send(chunk); err != nil {
			return fmt.Errorf("failed to send chunk: %v", err)
		}

		offset += int64(n)
		sentChunks++

		if sentChunks%10 == 0 {
			fmt.Printf("📦 Sent chunk %d: %d/%d bytes (%.1f%%)\n",
				sentChunks, offset, fileInfo.Size(),
				float64(offset)/float64(fileInfo.Size())*100)
		}
	}

	// Send last empty chunk to signal completion
	lastChunk := &proto.FileChunk{
		FileId:    fileID,
		Data:      []byte{},
		Offset:    offset,
		TotalSize: fileInfo.Size(),
		IsLast:    true,
	}
	stream.Send(lastChunk)

	fmt.Printf("✅ File download complete: %s (%d chunks, %d bytes)\n",
		fileID, sentChunks, offset)

	return nil
}
