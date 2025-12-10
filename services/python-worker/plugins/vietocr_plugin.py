"""
VietOCR Plugin - Vietnamese OCR using hybrid PaddleOCR + VietOCR approach
Uses PaddleOCR for text detection and VietOCR for Vietnamese text recognition
"""

import base64
import tempfile
import os
import numpy as np
import cv2
from datetime import datetime
from PIL import Image
from plugins.base_plugin import BasePlugin
from plugins.paddlet_onnx import paddlet_onnx

# VietOCR imports (lazy load to avoid startup delay)
_vietocr_predictor = None


def _get_vietocr_predictor():
    """Lazy load VietOCR predictor"""
    global _vietocr_predictor
    if _vietocr_predictor is None:
        try:
            from vietocr.tool.predictor import Predictor
            from vietocr.tool.config import Cfg
            
            # Use vgg_transformer config (good balance of speed/accuracy)
            config = Cfg.load_config_from_name('vgg_transformer')
            config['cnn']['pretrained'] = False  # Faster startup
            config['device'] = 'cpu'  # Use CPU for compatibility
            config['predictor']['beamsearch'] = False  # Faster inference
            
            _vietocr_predictor = Predictor(config)
            print("✓ VietOCR predictor initialized (vgg_transformer, CPU mode)")
        except Exception as e:
            print(f"⚠️  VietOCR initialization failed: {e}")
            print(f"    Fallback: Vietnamese OCR will use PaddleOCR only")
            _vietocr_predictor = None
    return _vietocr_predictor


class VietOCRPlugin(BasePlugin):
    """Vietnamese OCR using hybrid approach:
    - PaddleOCR for text detection (find bounding boxes)
    - VietOCR for text recognition (read Vietnamese text from boxes)
    
    This approach combines PaddleOCR's fast detection with VietOCR's
    superior Vietnamese text recognition accuracy.
    """
    
    def __init__(self):
        super().__init__()
        # Initialize PaddleOCR for detection
        weights_dir = os.path.join(os.path.dirname(__file__), 'weights')
        self.detection_engine = paddlet_onnx(weights_dir=weights_dir)
        print(f"✓ VietOCR Plugin: PaddleOCR detection engine ready")
    
    @property
    def name(self) -> str:
        return "vietocr_analyze"
    
    @property
    def description(self) -> str:
        return "Vietnamese OCR using PaddleOCR detection + VietOCR recognition"
    
    @property
    def input_schema(self) -> str:
        return '{"type":"object","properties":{"file":{"type":"string","format":"binary"}}}'
    
    @property
    def output_schema(self) -> str:
        return '{"type":"object","properties":{"texts":{"type":"array"},"confidences":{"type":"array"},"bboxes":{"type":"array"},"count":{"type":"number"},"filename":{"type":"string"},"engine":{"type":"string"},"status":{"type":"string"}}}'
    
    @property
    def accepts_file(self) -> bool:
        return True
    
    @property
    def file_field_name(self) -> str:
        return "file"
    
    def _process_with_vietocr(self, image_path: str) -> dict:
        """
        Hybrid OCR: PaddleOCR detection + VietOCR recognition
        
        Args:
            image_path: Path to image file
            
        Returns:
            Dict with texts, confidences, bboxes, count
        """
        # Load image
        image = cv2.imread(image_path)
        if image is None:
            raise ValueError(f"Cannot load image: {image_path}")
        
        # Convert BGR to RGB for processing
        rgb_image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        
        # Step 1: Use PaddleOCR Detection to find text regions
        # Call detection engine (it's a callable object)
        detection_boxes = self.detection_engine.detection(rgb_image)
        
        if detection_boxes is None or len(detection_boxes) == 0:
            return {
                'texts': [],
                'confidences': [],
                'bboxes': [],
                'count': 0,
                'engine': 'vietocr-hybrid'
            }
        
        # Step 2: Use VietOCR to recognize text in each bbox
        predictor = _get_vietocr_predictor()
        
        if predictor is None:
            # Fallback to full PaddleOCR if VietOCR failed to load
            print("⚠️  VietOCR unavailable, using PaddleOCR fallback")
            result = self.detection_engine.process_full_image(image_path)
            result['engine'] = 'paddleocr-fallback'
            return result
        
        texts = []
        confidences = []
        bboxes = []
        
        # Convert to PIL Image for VietOCR
        pil_image = Image.fromarray(rgb_image)
        
        for bbox_points in detection_boxes:
            # bbox_points is already a numpy array of shape (4, 2) with coordinates
            # Extract bbox coordinates
            pts = np.array(bbox_points, dtype=np.int32)
            
            # Get bounding rectangle
            x_min = int(np.min(pts[:, 0]))
            y_min = int(np.min(pts[:, 1]))
            x_max = int(np.max(pts[:, 0]))
            y_max = int(np.max(pts[:, 1]))
            
            # Crop region from PIL image
            try:
                cropped = pil_image.crop((x_min, y_min, x_max, y_max))
                
                # Recognize text with VietOCR
                text = predictor.predict(cropped)
                
                # VietOCR doesn't provide confidence, use 1.0 as placeholder
                texts.append(text)
                confidences.append(1.0)
                bboxes.append(bbox_points.tolist())  # Convert numpy to list for JSON
                
            except Exception as e:
                print(f"⚠️  Error processing bbox: {e}")
                continue
        
        return {
            'texts': texts,
            'confidences': confidences,
            'bboxes': bboxes,
            'count': len(texts),
            'engine': 'vietocr-hybrid'
        }
    
    def execute(self, params: dict, worker_sdk=None) -> dict:
        """Execute Vietnamese OCR on uploaded image
        
        Args:
            params: Dict containing:
                - file: base64 encoded image data (required)
                - filename: original filename (optional)
                - size: file size in bytes (optional)
        
        Returns:
            Dict with OCR results or error
        """
        filename = params.get('filename', 'unknown')
        size = params.get('size', 0)
        file_data = params.get('file')
        
        if not file_data:
            return {
                "error": "No file data provided",
                "status": "failed",
                "filename": filename
            }
        
        try:
            # Decode base64 file data
            image_bytes = base64.b64decode(file_data)
            
            # Save to temporary file
            with tempfile.NamedTemporaryFile(delete=False, suffix='.jpg') as tmp_file:
                tmp_file.write(image_bytes)
                tmp_file_path = tmp_file.name
            
            try:
                # Process with hybrid VietOCR
                print(f"🔍 VietOCR processing: {filename} ({size} bytes)")
                ocr_result = self._process_with_vietocr(tmp_file_path)
                
                # Convert numpy types to Python native types
                def convert_to_native(obj):
                    if isinstance(obj, np.ndarray):
                        return obj.tolist()
                    elif isinstance(obj, np.integer):
                        return int(obj)
                    elif isinstance(obj, np.floating):
                        return float(obj)
                    elif isinstance(obj, list):
                        return [convert_to_native(item) for item in obj]
                    elif isinstance(obj, dict):
                        return {key: convert_to_native(value) for key, value in obj.items()}
                    return obj
                
                ocr_result = convert_to_native(ocr_result)
                
                # Add metadata
                ocr_result['filename'] = filename
                ocr_result['size_bytes'] = size
                ocr_result['size_mb'] = round(size / (1024 * 1024), 2) if size > 0 else 0
                ocr_result['timestamp'] = datetime.now().isoformat()
                ocr_result['status'] = 'success'
                
                print(f"✓ VietOCR completed: Found {ocr_result.get('count', 0)} text regions")
                return ocr_result
                
            finally:
                # Clean up
                if os.path.exists(tmp_file_path):
                    os.unlink(tmp_file_path)
        
        except Exception as e:
            print(f"✗ VietOCR error: {e}")
            import traceback
            traceback.print_exc()
            return {
                "error": str(e),
                "status": "failed",
                "filename": filename,
                "size_bytes": size,
                "timestamp": datetime.now().isoformat()
            }
