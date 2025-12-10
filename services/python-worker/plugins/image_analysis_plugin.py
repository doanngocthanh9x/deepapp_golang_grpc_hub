"""
Image Analysis Plugin - Analyzes uploaded images using PaddleOCR
"""

import base64
import tempfile
import os
import numpy as np
from datetime import datetime
from plugins.base_plugin import BasePlugin
from plugins.paddlet_onnx import paddlet_onnx


class ImageAnalysisPlugin(BasePlugin):
    """Analyzes images using PaddleOCR and returns detected text"""
    
    def __init__(self):
        super().__init__()
        # Initialize PaddleOCR engine
        weights_dir = os.path.join(os.path.dirname(__file__), 'weights')
        self.ocr_engine = paddlet_onnx(weights_dir=weights_dir)
        print(f"✓ PaddleOCR engine initialized with weights from: {weights_dir}")
    
    @property
    def name(self) -> str:
        return "analyze_image"
    
    @property
    def description(self) -> str:
        return "Analyzes an image and returns information"
    
    @property
    def input_schema(self) -> str:
        return '{"type":"object","properties":{"file":{"type":"string","format":"binary"}}}'
    
    @property
    def output_schema(self) -> str:
        return '{"type":"object","properties":{"texts":{"type":"array"},"confidences":{"type":"array"},"bboxes":{"type":"array"},"count":{"type":"number"},"filename":{"type":"string"},"status":{"type":"string"}}}'
    
    @property
    def accepts_file(self) -> bool:
        return True
    
    @property
    def file_field_name(self) -> str:
        return "file"
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        """Analyze image file using PaddleOCR"""
        filename = params.get('filename', 'unknown')
        size = params.get('size', 0)
        file_data = params.get('file')  # base64 encoded file data
        
        if not file_data:
            return {
                "error": "No file data provided",
                "status": "failed",
                "filename": filename
            }
        
        try:
            # Decode base64 file data
            image_bytes = base64.b64decode(file_data)
            
            # Save to temporary file for OCR processing
            with tempfile.NamedTemporaryFile(delete=False, suffix='.jpg') as tmp_file:
                tmp_file.write(image_bytes)
                tmp_file_path = tmp_file.name
            
            try:
                # Process image with OCR
                print(f"🔍 Processing image: {filename} ({size} bytes)")
                ocr_result = self.ocr_engine.process_full_image(tmp_file_path)
                
                # Convert numpy types to Python native types for JSON serialization
                def convert_to_native(obj):
                    """Convert numpy types to Python native types"""
                    if isinstance(obj, np.ndarray):
                        return obj.tolist()
                    elif isinstance(obj, np.integer):  # Covers all int types
                        return int(obj)
                    elif isinstance(obj, np.floating):  # Covers all float types
                        return float(obj)
                    elif isinstance(obj, list):
                        return [convert_to_native(item) for item in obj]
                    elif isinstance(obj, dict):
                        return {key: convert_to_native(value) for key, value in obj.items()}
                    return obj
                
                # Convert all numpy types in result
                ocr_result = convert_to_native(ocr_result)
                
                # Add metadata
                ocr_result['filename'] = filename
                ocr_result['size_bytes'] = size
                ocr_result['size_mb'] = round(size / (1024 * 1024), 2) if size > 0 else 0
                ocr_result['timestamp'] = datetime.now().isoformat()
                ocr_result['status'] = 'success'
                
                print(f"✓ OCR completed: Found {ocr_result.get('count', 0)} text regions")
                return ocr_result
                
            finally:
                # Clean up temporary file
                if os.path.exists(tmp_file_path):
                    os.unlink(tmp_file_path)
        
        except Exception as e:
            print(f"✗ OCR error: {e}")
            return {
                "error": str(e),
                "status": "failed",
                "filename": filename,
                "size_bytes": size,
                "timestamp": datetime.now().isoformat()
            }

