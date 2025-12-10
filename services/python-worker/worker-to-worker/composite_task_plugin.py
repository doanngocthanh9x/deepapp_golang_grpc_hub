"""
Composite Plugin - Demonstrates Python worker calling other workers
"""
from plugins.base_plugin import BasePlugin
import json


class CompositeTaskPlugin(BasePlugin):
    """
    Composite task that calls multiple workers (Java, Node.js, Go)
    """

    def get_name(self):
        return "python_composite"

    def get_description(self):
        return "Calls multiple workers (Java, Node.js, Go) and combines their results"

    def execute(self, params, context):
        print("🔄 Starting composite task from Python...")

        results = {
            "worker_id": context["worker_id"],
            "timestamp": context.get("timestamp", ""),
            "calls": []
        }

        # Call Java worker
        try:
            print("  → Calling Java worker (hello_world)...")
            java_response = context["call_worker"](
                "java-simple-worker",
                "hello_world",
                {},
                timeout=10000
            )
            results["calls"].append({
                "worker": "java-simple-worker",
                "capability": "hello_world",
                "status": "success",
                "response": java_response
            })
        except Exception as e:
            print(f"  ✗ Java worker call failed: {e}")
            results["calls"].append({
                "worker": "java-simple-worker",
                "capability": "hello_world",
                "status": "error",
                "error": str(e)
            })

        # Call Node.js worker
        try:
            print("  → Calling Node.js worker (hello_node)...")
            node_response = context["call_worker"](
                "node-worker",
                "hello_node",
                {"name": "from Python"},
                timeout=10000
            )
            results["calls"].append({
                "worker": "node-worker",
                "capability": "hello_node",
                "status": "success",
                "response": node_response
            })
        except Exception as e:
            print(f"  ✗ Node.js worker call failed: {e}")
            results["calls"].append({
                "worker": "node-worker",
                "capability": "hello_node",
                "status": "error",
                "error": str(e)
            })

        # Call Go worker
        try:
            print("  → Calling Go worker (hello_go)...")
            go_response = context["call_worker"](
                "go-worker",
                "hello_go",
                {"name": "from Python"},
                timeout=10000
            )
            results["calls"].append({
                "worker": "go-worker",
                "capability": "hello_go",
                "status": "success",
                "response": go_response
            })
        except Exception as e:
            print(f"  ✗ Go worker call failed: {e}")
            results["calls"].append({
                "worker": "go-worker",
                "capability": "hello_go",
                "status": "error",
                "error": str(e)
            })

        # Summary
        success_count = sum(1 for call in results["calls"] if call["status"] == "success")
        results["summary"] = {
            "total_calls": len(results["calls"]),
            "successful": success_count,
            "failed": len(results["calls"]) - success_count
        }

        print("✅ Composite task completed")
        return results
