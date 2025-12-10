package plugins

import (
	"fmt"
	"time"
)

// HelloPlugin implements a simple hello world capability
type HelloPlugin struct {
	BasePlugin
}

// NewHelloGoPlugin creates a new HelloPlugin instance
func NewHelloGoPlugin() *HelloPlugin {
	return &HelloPlugin{}
}

func (p *HelloPlugin) GetName() string {
	return "hello_go"
}

func (p *HelloPlugin) GetDescription() string {
	return "Returns a hello message from Go worker"
}

func (p *HelloPlugin) Execute(params map[string]interface{}, context *ExecutionContext) (interface{}, error) {
	name := "World"
	if n, ok := params["name"].(string); ok {
		name = n
	}

	return map[string]interface{}{
		"message":    fmt.Sprintf("Hello %s from Go! 🔵", name),
		"worker_id":  context.WorkerID,
		"timestamp":  time.Now().Format(time.RFC3339),
		"go_version": "go1.18",
		"status":     "success",
	}, nil
}
