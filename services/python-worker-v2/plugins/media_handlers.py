"""
Image and media processing plugins
"""

import sys
import os
import base64
import io
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from decorators import capability
from PIL import Image
from vietocr.tool.predictor import Predictor
from vietocr.tool.config import Cfg


@capability(
    name="image_analysis",
    description="Analyze image and extract metadata",
    input_schema={
        "type": "object",
        "properties": {
            "image_url": {"type": "string", "description": "URL of image to analyze"},
            "operations": {
                "type": "array",
                "items": {"type": "string"},
                "description": "Operations to perform (metadata, colors, faces)"
            }
        },
        "required": ["image_url"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "width": {"type": "integer"},
            "height": {"type": "integer"},
            "format": {"type": "string"},
            "url": {"type": "string"},
            "analyzed_by": {"type": "string"}
        }
    }
)
def handle_image_analysis(worker_context, payload):
    """Analyze image and return metadata (mock implementation)"""
    image_url = payload.get("image_url", "")
    operations = payload.get("operations", ["metadata"])
    
    # Mock analysis
    result = {
        "width": 1920,
        "height": 1080,
        "format": "jpeg",
        "url": image_url,
        "operations_performed": operations,
        "analyzed_by": worker_context.get("worker_id")
    }
    
    if "colors" in operations:
        result["dominant_colors"] = ["#FF5733", "#33FF57", "#3357FF"]
    
    if "faces" in operations:
        result["faces_detected"] = 2
    
    return result


@capability(
    name="video_metadata",
    description="Extract metadata from video URL",
    input_schema={
        "type": "object",
        "properties": {
            "video_url": {"type": "string", "description": "URL of video"},
        },
        "required": ["video_url"]
    },
    output_schema={
        "type": "object",
        "properties": {
            "duration": {"type": "integer"},
            "resolution": {"type": "string"},
            "codec": {"type": "string"},
            "fps": {"type": "integer"}
        }
    }
)
def handle_video_metadata(worker_context, payload):
    """Extract video metadata (mock implementation)"""
    video_url = payload.get("video_url", "")
    
    return {
        "url": video_url,
        "duration": 120,  # seconds
        "resolution": "1920x1080",
        "codec": "h264",
        "fps": 30,
        "processed_by": worker_context.get("worker_id")
    }


# Initialize VietOCR predictor
def _get_vietocr_predictor():
    config = Cfg.load_config_from_name('vgg_transformer')
    config['device'] = 'cpu'  # Use CPU
    predictor = Predictor(config)
    return predictor


@capability(
    name="vietnamese_ocr",
    description="Extract Vietnamese text from image using VietOCR",
    input_schema={
        "type": "object",
        "properties": {
            "image_url": {"type": "string", "description": "URL of image to process"},
            "image_data": {"type": "string", "description": "Base64 encoded image data"},
            "language": {"type": "string", "enum": ["vi"], "default": "vi", "description": "Language for OCR"}
        },
        "oneOf": [
            {"required": ["image_url"]},
            {"required": ["image_data"]}
        ]
    },
    output_schema={
        "type": "object",
        "properties": {
            "text": {"type": "string", "description": "Extracted text"},
            "confidence": {"type": "number", "description": "OCR confidence score"},
            "language": {"type": "string", "description": "Detected language"},
            "processed_by": {"type": "string", "description": "Worker ID that processed the request"}
        },
        "required": ["text"]
    }
)
def handle_vietnamese_ocr(worker_context, payload):
    """Extract Vietnamese text from image using VietOCR"""
    try:
        # Get predictor (lazy initialization)
        if not hasattr(handle_vietnamese_ocr, '_predictor'):
            handle_vietnamese_ocr._predictor = _get_vietocr_predictor()
        
        predictor = handle_vietnamese_ocr._predictor
        
        # Get image
        image = None
        if "image_data" in payload:
            # Decode base64 image
            image_data = base64.b64decode(payload["image_data"])
            image = Image.open(io.BytesIO(image_data))
        elif "image_url" in payload:
            # For URL, we'd need to download, but for now assume local path or mock
            # In production, you'd use requests to download
            raise NotImplementedError("URL support not implemented yet")
        else:
            raise ValueError("Either image_url or image_data must be provided")
        
        # Convert to RGB if necessary
        if image.mode != 'RGB':
            image = image.convert('RGB')
        
        # Perform OCR
        text = predictor.predict(image)
        
        return {
            "text": text,
            "confidence": 0.95,  # VietOCR doesn't provide confidence, mock value
            "language": payload.get("language", "vi"),
            "processed_by": worker_context.get("worker_id")
        }
        
    except Exception as e:
        return {
            "error": str(e),
            "text": "",
            "processed_by": worker_context.get("worker_id")
        }
