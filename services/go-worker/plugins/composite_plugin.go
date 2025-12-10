package plugins

import "fmt"

// GoCompositePlugin demonstrates calling multiple operations
type GoCompositePlugin struct {
	BasePlugin
}

// NewGoCompositePlugin creates a new GoCompositePlugin instance
func NewGoCompositePlugin() *GoCompositePlugin {
	return &GoCompositePlugin{}
}

func (p *GoCompositePlugin) GetName() string {
	return "go_composite"
}

func (p *GoCompositePlugin) GetDescription() string {
	return "Composite operation: hash + base64 encode"
}

func (p *GoCompositePlugin) Execute(params map[string]interface{}, context *ExecutionContext) (interface{}, error) {
	text, ok := params["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("missing required parameter: text")
	}

	// This is a composite operation that combines hash and base64
	// In a real scenario, this could call other workers

	return map[string]interface{}{
		"input":     text,
		"operation": "composite (hash + base64)",
		"message":   fmt.Sprintf("Composite operation on: %s", text),
		"worker_id": context.WorkerID,
		"status":    "success",
	}, nil
}
