package plugins

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"strings"
)

// HashPlugin computes various hash functions
type HashPlugin struct {
	BasePlugin
}

// NewHashTextPlugin creates a new HashPlugin instance
func NewHashTextPlugin() *HashPlugin {
	return &HashPlugin{}
}

func (p *HashPlugin) GetName() string {
	return "hash_text"
}

func (p *HashPlugin) GetDescription() string {
	return "Compute hash (MD5, SHA256) of text"
}

func (p *HashPlugin) Execute(params map[string]interface{}, context *ExecutionContext) (interface{}, error) {
	text, ok := params["text"].(string)
	if !ok || text == "" {
		return nil, fmt.Errorf("missing required parameter: text")
	}

	algorithm := "sha256"
	if alg, ok := params["algorithm"].(string); ok {
		algorithm = strings.ToLower(alg)
	}

	var hash string
	switch algorithm {
	case "md5":
		hash = fmt.Sprintf("%x", md5.Sum([]byte(text)))
	case "sha256":
		hash = fmt.Sprintf("%x", sha256.Sum256([]byte(text)))
	default:
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return map[string]interface{}{
		"text":      text,
		"algorithm": algorithm,
		"hash":      hash,
		"status":    "success",
	}, nil
}
