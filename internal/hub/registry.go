package hub

import (
	"database/sql"
	"encoding/json"
	"sync"
	"time"
)

// ServiceCapability định nghĩa khả năng của một service
type ServiceCapability struct {
	Name          string `json:"name"`          // Tên capability (vd: "hello", "image_analysis")
	Description   string `json:"description"`   // Mô tả
	InputSchema   string `json:"input_schema,omitempty"`   // JSON schema string
	OutputSchema  string `json:"output_schema,omitempty"`  // JSON schema string
	HTTPMethod    string `json:"http_method"`               // GET, POST, PUT, DELETE
	AcceptsFile   bool   `json:"accepts_file"`              // Có nhận file upload không
	FileFieldName string `json:"file_field_name,omitempty"` // Tên field cho file
}

// WorkerInfo thông tin về worker
type WorkerInfo struct {
	ID           string               `json:"id"`
	Type         string               `json:"type"` // python, go, nodejs, etc
	Status       string               `json:"status"` // online, busy, offline
	Capabilities []ServiceCapability  `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	RegisteredAt string               `json:"registered_at"`
	LastSeen     string               `json:"last_seen"`
}

// ServiceRegistry quản lý workers và capabilities
type ServiceRegistry struct {
	mu            sync.RWMutex
	workers       map[string]*WorkerInfo              // worker_id -> info
	capabilities  map[string][]string                 // capability_name -> []worker_ids
	db            *sql.DB                             // Database connection
}

func NewServiceRegistry() *ServiceRegistry {
	return &ServiceRegistry{
		workers:      make(map[string]*WorkerInfo),
		capabilities: make(map[string][]string),
	}
}

func NewServiceRegistryWithDB(db *sql.DB) *ServiceRegistry {
	sr := &ServiceRegistry{
		workers:      make(map[string]*WorkerInfo),
		capabilities: make(map[string][]string),
		db:           db,
	}
	
	// Load existing workers from database on startup
	sr.loadFromDatabase()
	
	return sr
}

// loadFromDatabase loads workers and capabilities from database
func (sr *ServiceRegistry) loadFromDatabase() {
	if sr.db == nil {
		return
	}

	// Load workers
	rows, err := sr.db.Query(`
		SELECT id, type, status, metadata, registered_at, last_seen
		FROM workers WHERE status = 'online'
	`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var info WorkerInfo
		var metadataJSON sql.NullString
		
		err := rows.Scan(&info.ID, &info.Type, &info.Status, &metadataJSON,
			&info.RegisteredAt, &info.LastSeen)
		if err != nil {
			continue
		}

		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &info.Metadata)
		}

		// Load capabilities for this worker
		capRows, err := sr.db.Query(`
			SELECT name, description, input_schema, output_schema,
				http_method, accepts_file, file_field_name
			FROM capabilities WHERE worker_id = ?
		`, info.ID)
		if err != nil {
			continue
		}

		for capRows.Next() {
			var cap ServiceCapability
			var inputSchema, outputSchema, httpMethod, fileFieldName sql.NullString
			var acceptsFile sql.NullBool

			err := capRows.Scan(&cap.Name, &cap.Description, &inputSchema, &outputSchema,
				&httpMethod, &acceptsFile, &fileFieldName)
			if err != nil {
				continue
			}

			if inputSchema.Valid {
				cap.InputSchema = inputSchema.String
			}
			if outputSchema.Valid {
				cap.OutputSchema = outputSchema.String
			}
			if httpMethod.Valid {
				cap.HTTPMethod = httpMethod.String
			} else {
				cap.HTTPMethod = "POST"
			}
			if acceptsFile.Valid {
				cap.AcceptsFile = acceptsFile.Bool
			}
			if fileFieldName.Valid {
				cap.FileFieldName = fileFieldName.String
			}

			info.Capabilities = append(info.Capabilities, cap)
		}
		capRows.Close()

		sr.workers[info.ID] = &info

		// Index capabilities
		for _, cap := range info.Capabilities {
			sr.capabilities[cap.Name] = append(sr.capabilities[cap.Name], info.ID)
		}
	}
}

// RegisterWorker đăng ký worker với capabilities
func (sr *ServiceRegistry) RegisterWorker(workerID string, info *WorkerInfo) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	sr.workers[workerID] = info

	// Index capabilities in memory
	for _, cap := range info.Capabilities {
		if _, exists := sr.capabilities[cap.Name]; !exists {
			sr.capabilities[cap.Name] = []string{}
		}
		sr.capabilities[cap.Name] = append(sr.capabilities[cap.Name], workerID)
	}

	// Persist to database
	if sr.db != nil {
		sr.persistWorkerToDB(workerID, info)
	}
}

// persistWorkerToDB saves worker and capabilities to database
func (sr *ServiceRegistry) persistWorkerToDB(workerID string, info *WorkerInfo) {
	// Insert or update worker
	metadataJSON, _ := json.Marshal(info.Metadata)
	
	_, err := sr.db.Exec(`
		INSERT OR REPLACE INTO workers (id, type, status, metadata, registered_at, last_seen)
		VALUES (?, ?, ?, ?, ?, ?)
	`, workerID, info.Type, info.Status, string(metadataJSON), 
		time.Now(), time.Now())
	
	if err != nil {
		return
	}

	// Delete old capabilities
	sr.db.Exec(`DELETE FROM capabilities WHERE worker_id = ?`, workerID)

	// Insert new capabilities
	for _, cap := range info.Capabilities {
		sr.db.Exec(`
			INSERT INTO capabilities 
			(worker_id, name, description, input_schema, output_schema, http_method, accepts_file, file_field_name)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, workerID, cap.Name, cap.Description, cap.InputSchema, cap.OutputSchema,
			cap.HTTPMethod, cap.AcceptsFile, cap.FileFieldName)
	}
}

// UnregisterWorker gỡ đăng ký worker
func (sr *ServiceRegistry) UnregisterWorker(workerID string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	info, exists := sr.workers[workerID]
	if !exists {
		return
	}

	// Remove from capabilities index
	for _, cap := range info.Capabilities {
		workers := sr.capabilities[cap.Name]
		for i, wid := range workers {
			if wid == workerID {
				sr.capabilities[cap.Name] = append(workers[:i], workers[i+1:]...)
				break
			}
		}
	}

	delete(sr.workers, workerID)
}

// GetWorkerForCapability trả về worker ID có capability (load balancing đơn giản)
func (sr *ServiceRegistry) GetWorkerForCapability(capabilityName string) (string, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	workers, exists := sr.capabilities[capabilityName]
	if !exists || len(workers) == 0 {
		return "", false
	}

	// Simple round-robin: chọn worker đầu tiên có status online
	for _, workerID := range workers {
		if info, ok := sr.workers[workerID]; ok && info.Status == "online" {
			return workerID, true
		}
	}

	return "", false
}

// GetAllCapabilities trả về tất cả capabilities available
func (sr *ServiceRegistry) GetAllCapabilities() map[string]ServiceCapability {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make(map[string]ServiceCapability)
	
	for _, worker := range sr.workers {
		if worker.Status != "online" {
			continue
		}
		for _, cap := range worker.Capabilities {
			result[cap.Name] = cap
		}
	}

	return result
}

// GetAllWorkers trả về tất cả workers
func (sr *ServiceRegistry) GetAllWorkers() []*WorkerInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	workers := make([]*WorkerInfo, 0, len(sr.workers))
	for _, info := range sr.workers {
		workers = append(workers, info)
	}
	return workers
}

// UpdateWorkerStatus cập nhật status của worker
func (sr *ServiceRegistry) UpdateWorkerStatus(workerID, status string) {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if info, exists := sr.workers[workerID]; exists {
		info.Status = status
	}
}

// ToJSON serialize registry to JSON
func (sr *ServiceRegistry) ToJSON() ([]byte, error) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	data := map[string]interface{}{
		"workers":      sr.workers,
		"capabilities": sr.capabilities,
	}
	return json.Marshal(data)
}
