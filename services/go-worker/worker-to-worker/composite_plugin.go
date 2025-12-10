package workertoworker

import (
	"log"

	"github.com/deepapp/go-worker/plugins"
)

// CompositePlugin demonstrates worker-to-worker communication
type CompositePlugin struct {
	plugins.BasePlugin
}

func (p *CompositePlugin) GetName() string {
	return "go_composite"
}

func (p *CompositePlugin) GetDescription() string {
	return "Calls multiple workers (Python, Java, Node.js) and combines results"
}

func (p *CompositePlugin) Execute(params map[string]interface{}, context *plugins.ExecutionContext) (interface{}, error) {
	log.Println("🔄 Starting composite task...")

	results := map[string]interface{}{
		"worker_id": context.WorkerID,
		"calls":     []map[string]interface{}{},
	}

	calls := []map[string]interface{}{}

	// Call Python worker
	log.Println("  → Calling Python worker (hello)...")
	pythonResult, pythonErr := context.CallWorker("python-worker", "hello", map[string]interface{}{"name": "from Go"}, 10000)
	if pythonErr != nil {
		log.Printf("  ✗ Python worker call failed: %v", pythonErr)
		calls = append(calls, map[string]interface{}{
			"worker":     "python-worker",
			"capability": "hello",
			"status":     "error",
			"error":      pythonErr.Error(),
		})
	} else {
		calls = append(calls, map[string]interface{}{
			"worker":     "python-worker",
			"capability": "hello",
			"status":     "success",
			"response":   pythonResult,
		})
	}

	// Call Java worker
	log.Println("  → Calling Java worker (hello_world)...")
	javaResult, javaErr := context.CallWorker("java-simple-worker", "hello_world", map[string]interface{}{}, 10000)
	if javaErr != nil {
		log.Printf("  ✗ Java worker call failed: %v", javaErr)
		calls = append(calls, map[string]interface{}{
			"worker":     "java-simple-worker",
			"capability": "hello_world",
			"status":     "error",
			"error":      javaErr.Error(),
		})
	} else {
		calls = append(calls, map[string]interface{}{
			"worker":     "java-simple-worker",
			"capability": "hello_world",
			"status":     "success",
			"response":   javaResult,
		})
	}

	// Call Node.js worker
	log.Println("  → Calling Node.js worker (hello_node)...")
	nodeResult, nodeErr := context.CallWorker("node-worker", "hello_node", map[string]interface{}{"name": "from Go"}, 10000)
	if nodeErr != nil {
		log.Printf("  ✗ Node.js worker call failed: %v", nodeErr)
		calls = append(calls, map[string]interface{}{
			"worker":     "node-worker",
			"capability": "hello_node",
			"status":     "error",
			"error":      nodeErr.Error(),
		})
	} else {
		calls = append(calls, map[string]interface{}{
			"worker":     "node-worker",
			"capability": "hello_node",
			"status":     "success",
			"response":   nodeResult,
		})
	}

	results["calls"] = calls

	// Summary
	successCount := 0
	for _, call := range calls {
		if call["status"] == "success" {
			successCount++
		}
	}

	results["summary"] = map[string]interface{}{
		"total_calls": len(calls),
		"successful":  successCount,
		"failed":      len(calls) - successCount,
	}

	log.Println("✅ Composite task completed")
	return results, nil
}
