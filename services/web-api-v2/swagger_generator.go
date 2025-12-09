package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SwaggerGenerator generates OpenAPI spec from worker capabilities
type SwaggerGenerator struct {
	capabilities map[string]interface{}
}

func NewSwaggerGenerator() *SwaggerGenerator {
	return &SwaggerGenerator{
		capabilities: make(map[string]interface{}),
	}
}

func (sg *SwaggerGenerator) UpdateCapabilities(caps map[string]interface{}) {
	sg.capabilities = caps
}

func (sg *SwaggerGenerator) GenerateSpec() map[string]interface{} {
	spec := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"title":       "gRPC Hub Dynamic API",
			"description": "Auto-generated API from worker capabilities",
			"version":     "2.0",
		},
		"host":     "localhost:8082",
		"basePath": "/api/v2",
		"schemes":  []string{"http"},
		"paths":    sg.generatePaths(),
		"definitions": map[string]interface{}{
			"Error": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"error": map[string]interface{}{
						"type": "string",
					},
				},
			},
		},
	}
	
	return spec
}

func (sg *SwaggerGenerator) generatePaths() map[string]interface{} {
	paths := make(map[string]interface{})
	
	// Static endpoints
	paths["/status"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"system"},
			"summary":     "Get API status",
			"description": "Get current API gateway status and info",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
					"schema": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"client_id": map[string]interface{}{"type": "string"},
							"status":    map[string]interface{}{"type": "string"},
							"version":   map[string]interface{}{"type": "string"},
							"timestamp": map[string]interface{}{"type": "string"},
						},
					},
				},
			},
		},
	}
	
	paths["/capabilities"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{"discovery"},
			"summary":     "List all capabilities",
			"description": "Get all service capabilities registered by workers",
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "Successful response",
				},
			},
		},
	}
	
	// Dynamic endpoints from capabilities
	for capName, capInfoRaw := range sg.capabilities {
		capInfo, ok := capInfoRaw.(map[string]interface{})
		if !ok {
			continue
		}
		
		path := fmt.Sprintf("/invoke/%s", capName)
		description := "No description"
		if desc, ok := capInfo["description"].(string); ok {
			description = desc
		}
		
		// Parse input schema
		var inputSchema map[string]interface{}
		if inputSchemaStr, ok := capInfo["input_schema"].(string); ok {
			json.Unmarshal([]byte(inputSchemaStr), &inputSchema)
		}
		
		// Parse output schema
		var outputSchema map[string]interface{}
		if outputSchemaStr, ok := capInfo["output_schema"].(string); ok {
			json.Unmarshal([]byte(outputSchemaStr), &outputSchema)
		}
		
		// Generate tag from capability name
		tag := sg.generateTag(capName)
		
		paths[path] = map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{tag},
				"summary":     strings.Title(strings.ReplaceAll(capName, "_", " ")),
				"description": description,
				"consumes":    []string{"application/json"},
				"produces":    []string{"application/json"},
				"parameters": []map[string]interface{}{
					{
						"in":          "body",
						"name":        "body",
						"description": "Request payload",
						"required":    true,
						"schema":      sg.convertSchema(inputSchema),
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful response",
						"schema":      sg.convertSchema(outputSchema),
					},
					"404": map[string]interface{}{
						"description": "Capability not found",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/Error",
						},
					},
					"500": map[string]interface{}{
						"description": "Internal server error",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/Error",
						},
					},
					"408": map[string]interface{}{
						"description": "Request timeout",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/Error",
						},
					},
				},
			},
		}
	}
	
	return paths
}

func (sg *SwaggerGenerator) generateTag(capName string) string {
	// Group by prefix
	if strings.HasPrefix(capName, "text_") {
		return "text"
	}
	if strings.HasPrefix(capName, "image_") || strings.HasPrefix(capName, "video_") {
		return "media"
	}
	if strings.Contains(capName, "calc") || strings.Contains(capName, "math") || strings.Contains(capName, "stat") {
		return "math"
	}
	return "services"
}

func (sg *SwaggerGenerator) convertSchema(schema map[string]interface{}) map[string]interface{} {
	if schema == nil || len(schema) == 0 {
		return map[string]interface{}{
			"type": "object",
		}
	}
	
	// Already valid JSON Schema, return as-is
	return schema
}

func (sg *SwaggerGenerator) GetJSON() ([]byte, error) {
	spec := sg.GenerateSpec()
	return json.MarshalIndent(spec, "", "  ")
}
