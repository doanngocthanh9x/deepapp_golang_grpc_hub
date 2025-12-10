package plugins

import (
	"encoding/base64"
	"fmt"
)

// Base64Plugin encodes/decodes base64
type Base64Plugin struct {
	BasePlugin
}

// NewBase64OpsPlugin creates a new Base64Plugin instance
func NewBase64OpsPlugin() *Base64Plugin {
	return &Base64Plugin{}
}

func (p *Base64Plugin) GetName() string {
	return "base64_ops"
}

func (p *Base64Plugin) GetDescription() string {
	return "Encode or decode base64 strings"
}

func (p *Base64Plugin) Execute(params map[string]interface{}, context *ExecutionContext) (interface{}, error) {
	text, ok := params["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("missing required parameter: text")
	}

	operation := "encode"
	if op, ok := params["operation"].(string); ok {
		operation = op
	}

	var result string
	var err error

	switch operation {
	case "encode":
		result = base64.StdEncoding.EncodeToString([]byte(text))
	case "decode":
		decoded, decodeErr := base64.StdEncoding.DecodeString(text)
		if decodeErr != nil {
			return nil, fmt.Errorf("decode error: %v", decodeErr)
		}
		result = string(decoded)
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"input":     text,
		"operation": operation,
		"result":    result,
		"status":    "success",
	}, nil
}
