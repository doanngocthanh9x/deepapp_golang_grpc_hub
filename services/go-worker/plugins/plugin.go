package plugins

import "encoding/json"

// Plugin represents a capability that can be executed
type Plugin interface {
	// GetName returns the capability name
	GetName() string

	// GetDescription returns the capability description
	GetDescription() string

	// GetHttpMethod returns the HTTP method (GET, POST, etc.)
	GetHttpMethod() string

	// AcceptsFile returns true if this capability accepts file uploads
	AcceptsFile() bool

	// GetFileFieldName returns the field name for file uploads
	GetFileFieldName() string

	// Execute runs the plugin logic
	Execute(params map[string]interface{}, context *ExecutionContext) (interface{}, error)
}

// ExecutionContext provides context for plugin execution
type ExecutionContext struct {
	WorkerID   string
	CallWorker CallWorkerFunc
}

// CallWorkerFunc is a function type for calling other workers
type CallWorkerFunc func(targetWorkerID, capability string, params map[string]interface{}, timeout int) (map[string]interface{}, error)

// BasePlugin provides default implementations
type BasePlugin struct{}

func (b *BasePlugin) GetHttpMethod() string {
	return "POST"
}

func (b *BasePlugin) AcceptsFile() bool {
	return false
}

func (b *BasePlugin) GetFileFieldName() string {
	return "file"
}

// Capability represents capability metadata for registration
type Capability struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	HttpMethod    string `json:"http_method"`
	AcceptsFile   bool   `json:"accepts_file"`
	FileFieldName string `json:"file_field_name"`
}

// ToCapability converts a Plugin to Capability metadata
func ToCapability(p Plugin) *Capability {
	return &Capability{
		Name:          p.GetName(),
		Description:   p.GetDescription(),
		HttpMethod:    p.GetHttpMethod(),
		AcceptsFile:   p.AcceptsFile(),
		FileFieldName: p.GetFileFieldName(),
	}
}

// ParseParams parses JSON data into a map
func ParseParams(data string) (map[string]interface{}, error) {
	var params map[string]interface{}
	if data == "" {
		return make(map[string]interface{}), nil
	}
	err := json.Unmarshal([]byte(data), &params)
	return params, err
}

// FormatResult formats result as JSON string
func FormatResult(result interface{}) (string, error) {
	data, err := json.Marshal(result)
	return string(data), err
}
