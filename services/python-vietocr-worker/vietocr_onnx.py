#!/usr/bin/env python3
"""
VietOCR ONNX Inference Module
High-performance OCR using ONNX Runtime
"""

import onnxruntime as ort
import numpy as np
import cv2
from typing import List, Tuple, Optional, Dict
import yaml
import base64
from io import BytesIO
from PIL import Image
import os


class VietOCR_ONNX:
    """VietOCR inference with ONNX Runtime"""
    
    def __init__(
        self, 
        encoder_path: str,
        decoder_path: str,
        config_path: Optional[str] = None,
        vocab_path: Optional[str] = None,
        use_gpu: bool = False
    ):
        """
        Initialize VietOCR ONNX inference
        
        Args:
            encoder_path: Path to encoder ONNX model
            decoder_path: Path to decoder ONNX model
            config_path: Path to config.yml (optional)
            vocab_path: Path to vocab.txt (optional)
            use_gpu: Use GPU acceleration
        """
        # Get execution providers
        providers = self._get_providers(use_gpu)
        
        # Load ONNX sessions
        self.encoder_session = ort.InferenceSession(
            encoder_path, 
            providers=providers
        )
        self.decoder_session = ort.InferenceSession(
            decoder_path,
            providers=providers
        )
        
        # Load config
        self.config = self._load_config(config_path)
        
        # Load vocabulary
        self.vocab = self._load_vocab(vocab_path)
        self.idx2char = {i: char for i, char in enumerate(self.vocab)}
        self.char2idx = {char: i for i, char in enumerate(self.vocab)}
        
        # Special tokens
        self.sos_token = self.char2idx.get('<sos>', 1)
        self.eos_token = self.char2idx.get('<eos>', 2)
        self.pad_token = self.char2idx.get('<pad>', 0)
        
        print(f"✓ VietOCR ONNX initialized")
        print(f"  - Encoder: {encoder_path}")
        print(f"  - Decoder: {decoder_path}")
        print(f"  - Vocab size: {len(self.vocab)}")
        print(f"  - Providers: {providers}")
    
    def _get_providers(self, use_gpu: bool) -> List[str]:
        """Get ONNX Runtime execution providers"""
        available = ort.get_available_providers()
        
        if use_gpu:
            if 'CUDAExecutionProvider' in available:
                return ['CUDAExecutionProvider', 'CPUExecutionProvider']
            elif 'TensorrtExecutionProvider' in available:
                return ['TensorrtExecutionProvider', 'CUDAExecutionProvider', 'CPUExecutionProvider']
        
        return ['CPUExecutionProvider']
    
    def _load_config(self, config_path: Optional[str]) -> Dict:
        """Load configuration"""
        if config_path and os.path.exists(config_path):
            with open(config_path, 'r') as f:
                return yaml.safe_load(f)
        
        # Default config
        return {
            'image_height': 32,
            'image_min_width': 32,
            'image_max_width': 512
        }
    
    def _load_vocab(self, vocab_path: Optional[str]) -> List[str]:
        """Load vocabulary"""
        if vocab_path and os.path.exists(vocab_path):
            with open(vocab_path, 'r', encoding='utf-8') as f:
                vocab = [line.strip() for line in f]
        else:
            # Default Vietnamese vocab
            vocab = ['<pad>', '<sos>', '<eos>']
            vocab += list('aAàÀảẢãÃáÁạẠăĂằẰẳẲẵẴắẮặẶâÂầầẦẩẨẫẪấẤậẬbBcCdDđĐeEèÈẻẺẽẼéÉẹẸêÊềỀểỂễỄếẾệỆfFgGhHiIìÌỉỈĩĨíÍịỊjJkKlLmMnNoOòÒỏỎõÕóÓọỌôÔồỒổỔỗỖốỐộỘơƠờỜởỞỡỠớỚợỢpPqQrRsStTuUùÙủỦũŨúÚụỤưƯừỪửỬữỮứỨựỰvVwWxXyYỳỲỷỶỹỸýÝỵỴzZ')
            vocab += list('0123456789!"#$%&\'()*+,-./:;<=>?@[\\]^_`{|}~ ')
        
        return vocab
    
    def preprocess_image(
        self, 
        image_input,
        target_height: int = 32
    ) -> np.ndarray:
        """
        Preprocess image for OCR
        
        Args:
            image_input: Image path, numpy array, PIL Image, or base64 string
            target_height: Target height
            
        Returns:
            Preprocessed image tensor (B, C, H, W)
        """
        # Handle different input types
        if isinstance(image_input, str):
            if image_input.startswith('data:image') or image_input.startswith('/9j/') or image_input.startswith('iVBOR'):
                # Base64 encoded image
                if 'base64,' in image_input:
                    image_input = image_input.split('base64,')[1]
                img_data = base64.b64decode(image_input)
                img = Image.open(BytesIO(img_data))
                img = np.array(img)
            else:
                # File path
                img = cv2.imread(image_input)
        elif isinstance(image_input, Image.Image):
            img = np.array(image_input)
        else:
            img = image_input
        
        if img is None:
            raise ValueError("Cannot read image")
        
        # Convert to RGB
        if len(img.shape) == 2:
            img = cv2.cvtColor(img, cv2.COLOR_GRAY2RGB)
        elif img.shape[2] == 4:
            img = cv2.cvtColor(img, cv2.COLOR_BGRA2RGB)
        elif len(img.shape) == 3 and img.shape[2] == 3:
            # Check if BGR or RGB
            if isinstance(image_input, str) and not image_input.startswith('data:'):
                img = cv2.cvtColor(img, cv2.COLOR_BGR2RGB)
        
        # Resize maintaining aspect ratio
        h, w = img.shape[:2]
        aspect_ratio = w / h
        target_width = int(target_height * aspect_ratio)
        
        # Clamp width
        min_width = self.config.get('image_min_width', 32)
        max_width = self.config.get('image_max_width', 512)
        target_width = max(min_width, min(target_width, max_width))
        
        img = cv2.resize(img, (target_width, target_height))
        
        # Normalize to [0, 1]
        img = img.astype(np.float32) / 255.0
        
        # Transpose (H, W, C) -> (C, H, W)
        img = np.transpose(img, (2, 0, 1))
        
        # Add batch dimension
        img = np.expand_dims(img, axis=0)
        
        return img
    
    def predict(
        self,
        image_input,
        max_seq_length: int = 128,
        return_prob: bool = False
    ) -> Tuple[str, float]:
        """
        Predict text from image
        
        Args:
            image_input: Image (path, array, PIL, or base64)
            max_seq_length: Maximum sequence length
            return_prob: Return confidence probability
            
        Returns:
            Predicted text and confidence
        """
        # Preprocess
        img = self.preprocess_image(image_input)
        
        # Encoder forward
        encoder_input = {self.encoder_session.get_inputs()[0].name: img}
        memory = self.encoder_session.run(None, encoder_input)[0]
        
        # Decoder autoregressive generation
        tgt_inp = np.array([[self.sos_token]], dtype=np.int64)
        translated_indices = []
        probs = []
        
        for step in range(max_seq_length):
            # Decoder forward
            decoder_input = {
                self.decoder_session.get_inputs()[0].name: tgt_inp,
                self.decoder_session.get_inputs()[1].name: memory
            }
            
            output = self.decoder_session.run(None, decoder_input)[0]
            
            # Get last token prediction
            last_output = output[-1, 0, :]
            
            # Softmax
            exp_output = np.exp(last_output - np.max(last_output))
            probabilities = exp_output / np.sum(exp_output)
            
            # Greedy decoding
            next_token = np.argmax(probabilities)
            prob = probabilities[next_token]
            
            # Stop if EOS
            if next_token == self.eos_token:
                break
            
            translated_indices.append(next_token)
            probs.append(prob)
            
            # Update input
            tgt_inp = np.concatenate([tgt_inp, [[next_token]]], axis=0)
        
        # Decode to text
        text = ''.join([
            self.idx2char.get(int(idx), '') 
            for idx in translated_indices
        ])
        
        avg_prob = np.mean(probs) if probs else 0.0
        
        if return_prob:
            return text, float(avg_prob)
        
        return text, 1.0
    
    def predict_batch(
        self,
        image_inputs: List,
        max_seq_length: int = 128
    ) -> List[Tuple[str, float]]:
        """
        Batch inference
        
        Args:
            image_inputs: List of images
            max_seq_length: Maximum sequence length
            
        Returns:
            List of (text, confidence) tuples
        """
        results = []
        for img_input in image_inputs:
            try:
                text, prob = self.predict(img_input, max_seq_length, return_prob=True)
                results.append((text, prob))
            except Exception as e:
                print(f"Error processing image: {e}")
                results.append(("", 0.0))
        
        return results
