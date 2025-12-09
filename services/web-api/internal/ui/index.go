package ui

import (
	"fmt"
	"net/http"
)

// IndexHandler handles the main UI page
type IndexHandler struct{}

// NewIndexHandler creates a new index handler
func NewIndexHandler() *IndexHandler {
	return &IndexHandler{}
}

// HandleIndex serves the main UI page
func (h *IndexHandler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>DeepApp gRPC Hub - Web API</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            overflow: hidden;
        }

        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 40px;
            text-align: center;
        }

        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }

        .header p {
            font-size: 1.2em;
            opacity: 0.9;
        }

        .content {
            padding: 40px;
        }

        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 30px;
            border-bottom: 2px solid #e0e0e0;
        }

        .tab {
            padding: 15px 30px;
            background: none;
            border: none;
            cursor: pointer;
            font-size: 1.1em;
            color: #666;
            transition: all 0.3s;
            border-bottom: 3px solid transparent;
        }

        .tab.active {
            color: #667eea;
            border-bottom-color: #667eea;
        }

        .tab-content {
            display: none;
        }

        .tab-content.active {
            display: block;
        }

        .endpoint-card {
            background: #f8f9fa;
            border-radius: 15px;
            padding: 25px;
            margin-bottom: 20px;
            box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
            transition: transform 0.3s, box-shadow 0.3s;
        }

        .endpoint-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 5px 20px rgba(0, 0, 0, 0.15);
        }

        .endpoint-card h3 {
            color: #667eea;
            margin-bottom: 15px;
            font-size: 1.5em;
        }

        .endpoint-card .method {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 5px;
            font-size: 0.9em;
            font-weight: bold;
            margin-bottom: 10px;
        }

        .method.GET {
            background: #4CAF50;
            color: white;
        }

        .method.POST {
            background: #2196F3;
            color: white;
        }

        .endpoint-path {
            font-family: 'Courier New', monospace;
            background: white;
            padding: 10px 15px;
            border-radius: 5px;
            margin: 10px 0;
            border-left: 4px solid #667eea;
        }

        .input-group {
            margin: 15px 0;
        }

        .input-group label {
            display: block;
            margin-bottom: 5px;
            color: #555;
            font-weight: 500;
        }

        .input-group input,
        .input-group textarea {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 1em;
            transition: border-color 0.3s;
        }

        .input-group input:focus,
        .input-group textarea:focus {
            outline: none;
            border-color: #667eea;
        }

        button.test-btn {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 12px 30px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 1.1em;
            font-weight: bold;
            transition: transform 0.3s, box-shadow 0.3s;
        }

        button.test-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
        }

        button.test-btn:active {
            transform: translateY(0);
        }

        #result {
            background: #f0f4f8;
            border-radius: 10px;
            padding: 20px;
            margin-top: 30px;
            display: none;
            border-left: 5px solid #4CAF50;
        }

        #result.show {
            display: block;
            animation: slideIn 0.3s ease-out;
        }

        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        #result h3 {
            color: #4CAF50;
            margin-bottom: 15px;
        }

        #result pre {
            background: white;
            padding: 15px;
            border-radius: 5px;
            overflow-x: auto;
            font-family: 'Courier New', monospace;
            font-size: 0.95em;
            line-height: 1.5;
        }

        .quick-links {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }

        .quick-link {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 20px;
            border-radius: 10px;
            text-align: center;
            text-decoration: none;
            transition: transform 0.3s;
        }

        .quick-link:hover {
            transform: scale(1.05);
        }

        .quick-link h4 {
            margin-bottom: 10px;
            font-size: 1.3em;
        }

        .loader {
            border: 4px solid #f3f3f3;
            border-top: 4px solid #667eea;
            border-radius: 50%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 20px auto;
            display: none;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 DeepApp gRPC Hub</h1>
            <p>Web API Gateway for Microservices</p>
        </div>

        <div class="content">
            <div class="tabs">
                <button class="tab active" onclick="showTab('python')">🐍 Python Worker</button>
                <button class="tab" onclick="showTab('java')">☕ Java Worker</button>
                <button class="tab" onclick="showTab('dynamic')">⚡ Dynamic API</button>
                <button class="tab" onclick="showTab('info')">ℹ️ Info</button>
            </div>

            <!-- Python Worker Tab -->
            <div id="python-tab" class="tab-content active">
                <div class="endpoint-card">
                    <h3>Hello World</h3>
                    <span class="method GET">GET</span>
                    <span class="method POST">POST</span>
                    <div class="endpoint-path">/api/worker/python/hello</div>
                    <p>Simple hello world response from Python worker</p>
                    <button class="test-btn" onclick="testPythonHello()">Test Hello</button>
                </div>

                <div class="endpoint-card">
                    <h3>Analyze Image</h3>
                    <span class="method POST">POST</span>
                    <div class="endpoint-path">/api/worker/python/analyze_image</div>
                    <p>Upload and analyze an image</p>
                    <div class="input-group">
                        <label>Select Image:</label>
                        <input type="file" id="pythonImageFile" accept="image/*">
                    </div>
                    <button class="test-btn" onclick="testPythonAnalyze()">Analyze Image</button>
                </div>
            </div>

            <!-- Java Worker Tab -->
            <div id="java-tab" class="tab-content">
                <div class="endpoint-card">
                    <h3>Hello World</h3>
                    <span class="method GET">GET</span>
                    <span class="method POST">POST</span>
                    <div class="endpoint-path">/api/worker/java/hello</div>
                    <p>Simple hello world response from Java worker</p>
                    <button class="test-btn" onclick="testJavaHello()">Test Hello</button>
                </div>

                <div class="endpoint-card">
                    <h3>File Info</h3>
                    <span class="method POST">POST</span>
                    <div class="endpoint-path">/api/worker/java/file_info</div>
                    <p>Get information about a file on the server</p>
                    <div class="input-group">
                        <label>File Path:</label>
                        <input type="text" id="javaFilePath" placeholder="/etc/hosts" value="/etc/hosts">
                    </div>
                    <button class="test-btn" onclick="testJavaFileInfo()">Get File Info</button>
                </div>
            </div>

            <!-- Dynamic API Tab -->
            <div id="dynamic-tab" class="tab-content">
                <div class="endpoint-card">
                    <h3>Capabilities Discovery</h3>
                    <span class="method GET">GET</span>
                    <div class="endpoint-path">/api/capabilities</div>
                    <p>Get all available worker capabilities</p>
                    <button class="test-btn" onclick="testCapabilities()">Get Capabilities</button>
                </div>

                <div class="endpoint-card">
                    <h3>Swagger API Documentation</h3>
                    <span class="method GET">GET</span>
                    <div class="endpoint-path">/api/swagger.json</div>
                    <p>OpenAPI 3.0 specification</p>
                    <button class="test-btn" onclick="window.open('/api/swagger.json', '_blank')">View Swagger JSON</button>
                </div>

                <div class="endpoint-card">
                    <h3>Interactive API Docs</h3>
                    <span class="method GET">GET</span>
                    <div class="endpoint-path">/api/docs</div>
                    <p>Swagger UI for testing APIs</p>
                    <button class="test-btn" onclick="window.open('/api/docs', '_blank')">Open Swagger UI</button>
                </div>

                <div class="endpoint-card">
                    <h3>Dynamic Call</h3>
                    <span class="method POST">POST</span>
                    <div class="endpoint-path">/api/call/{capability}</div>
                    <p>Call any registered capability dynamically</p>
                    <div class="input-group">
                        <label>Capability Name:</label>
                        <input type="text" id="dynamicCapability" placeholder="hello" value="hello">
                    </div>
                    <div class="input-group">
                        <label>Request Data (JSON):</label>
                        <textarea id="dynamicData" rows="4" placeholder='{"key": "value"}'>{}</textarea>
                    </div>
                    <button class="test-btn" onclick="testDynamicCall()">Execute Call</button>
                </div>
            </div>

            <!-- Info Tab -->
            <div id="info-tab" class="tab-content">
                <div class="endpoint-card">
                    <h3>System Status</h3>
                    <span class="method GET">GET</span>
                    <div class="endpoint-path">/api/status</div>
                    <p>Check API health and available endpoints</p>
                    <button class="test-btn" onclick="testStatus()">Check Status</button>
                </div>

                <div class="quick-links">
                    <a href="/api/docs" target="_blank" class="quick-link">
                        <h4>📚 API Docs</h4>
                        <p>Interactive Swagger UI</p>
                    </a>
                    <a href="/api/capabilities" target="_blank" class="quick-link">
                        <h4>🔍 Capabilities</h4>
                        <p>View all workers</p>
                    </a>
                    <a href="/api/status" target="_blank" class="quick-link">
                        <h4>💚 Status</h4>
                        <p>System health</p>
                    </a>
                    <a href="/api/swagger.json" target="_blank" class="quick-link">
                        <h4>📄 OpenAPI Spec</h4>
                        <p>Swagger JSON</p>
                    </a>
                </div>
            </div>

            <div class="loader" id="loader"></div>
            <div id="result"></div>
        </div>
    </div>

    <script>
        function showTab(tabName) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            document.querySelectorAll('.tab').forEach(btn => {
                btn.classList.remove('active');
            });

            // Show selected tab
            document.getElementById(tabName + '-tab').classList.add('active');
            event.target.classList.add('active');

            // Clear result
            document.getElementById('result').classList.remove('show');
        }

        function showLoader() {
            document.getElementById('loader').style.display = 'block';
            document.getElementById('result').classList.remove('show');
        }

        function hideLoader() {
            document.getElementById('loader').style.display = 'none';
        }

        function showResult(data) {
            hideLoader();
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = '<h3>✅ Response:</h3><pre>' + JSON.stringify(data, null, 2) + '</pre>';
            resultDiv.classList.add('show');
        }

        function showError(error) {
            hideLoader();
            const resultDiv = document.getElementById('result');
            resultDiv.innerHTML = '<h3>❌ Error:</h3><pre>' + error + '</pre>';
            resultDiv.style.borderLeftColor = '#f44336';
            resultDiv.classList.add('show');
        }

        // Python Worker Functions
        function testPythonHello() {
            showLoader();
            fetch('/api/worker/python/hello', { method: 'POST' })
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }

        function testPythonAnalyze() {
            const file = document.getElementById('pythonImageFile').files[0];
            if (!file) {
                alert('Please select an image');
                return;
            }

            const formData = new FormData();
            formData.append('image', file);

            showLoader();
            fetch('/api/worker/python/analyze_image', { method: 'POST', body: formData })
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }

        // Java Worker Functions
        function testJavaHello() {
            showLoader();
            fetch('/api/worker/java/hello', { method: 'POST' })
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }

        function testJavaFileInfo() {
            const filePath = document.getElementById('javaFilePath').value;
            if (!filePath) {
                alert('Please enter a file path');
                return;
            }

            showLoader();
            fetch('/api/worker/java/file_info', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ filePath: filePath })
            })
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }

        // Dynamic API Functions
        function testCapabilities() {
            showLoader();
            fetch('/api/capabilities')
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }

        function testDynamicCall() {
            const capability = document.getElementById('dynamicCapability').value;
            const data = document.getElementById('dynamicData').value;

            if (!capability) {
                alert('Please enter a capability name');
                return;
            }

            let body;
            try {
                body = JSON.parse(data);
            } catch (e) {
                alert('Invalid JSON data');
                return;
            }

            showLoader();
            fetch('/api/call/' + capability, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(body)
            })
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }

        function testStatus() {
            showLoader();
            fetch('/api/status')
                .then(r => r.json())
                .then(showResult)
                .catch(err => showError(err.message));
        }
    </script>
</body>
</html>
	`)
}