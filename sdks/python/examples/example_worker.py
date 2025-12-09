#!/usr/bin/env python3
"""
Example Python worker using the DeepApp SDK
"""

import os
import json
import base64
from datetime import datetime
from deepapp_sdk import Worker


class ExampleWorker(Worker):
    """Example worker demonstrating various capabilities"""

    def __init__(self):
        super().__init__(
            worker_id=os.getenv('WORKER_ID', 'python-example-worker'),
            hub_address=os.getenv('HUB_ADDRESS', 'localhost:50051')
        )

    def get_capabilities(self):
        """Define the capabilities this worker provides"""
        return [
            {
                'name': 'hello',
                'description': 'Returns a hello message',
                'input_schema': '{}',
                'output_schema': '{"type":"object","properties":{"message":{"type":"string"},"timestamp":{"type":"string"},"worker_id":{"type":"string"},"status":{"type":"string"}}}',
                'http_method': 'GET',
                'accepts_file': False
            },
            {
                'name': 'echo',
                'description': 'Echoes back the input message',
                'input_schema': '{"type":"object","properties":{"message":{"type":"string"}}}',
                'output_schema': '{"type":"object","properties":{"echo":{"type":"string"},"timestamp":{"type":"string"},"status":{"type":"string"}}}',
                'http_method': 'POST',
                'accepts_file': False
            },
            {
                'name': 'reverse_text',
                'description': 'Reverses the input text',
                'input_schema': '{"type":"object","properties":{"text":{"type":"string"}}}',
                'output_schema': '{"type":"object","properties":{"original":{"type":"string"},"reversed":{"type":"string"},"timestamp":{"type":"string"},"status":{"type":"string"}}}',
                'http_method': 'POST',
                'accepts_file': False
            },
            {
                'name': 'analyze_file',
                'description': 'Analyze an uploaded file',
                'input_schema': '{"type":"object","properties":{"file":{"type":"string","format":"binary"},"filename":{"type":"string"}}}',
                'output_schema': '{"type":"object","properties":{"filename":{"type":"string"},"size":{"type":"number"},"mime_type":{"type":"string"},"analysis":{"type":"object"},"timestamp":{"type":"string"},"status":{"type":"string"}}}',
                'http_method': 'POST',
                'accepts_file': True,
                'file_field_name': 'file'
            }
        ]

    def handle_hello(self, message):
        """Handle hello capability"""
        print("🔍 Processing hello request")

        return {
            'message': 'Hello World from Python SDK Worker! 🐍',
            'timestamp': datetime.now().isoformat(),
            'worker_id': self.worker_id,
            'status': 'success'
        }

    def handle_echo(self, message):
        """Handle echo capability"""
        print("🔍 Processing echo request")

        try:
            content = json.loads(message.content)
            input_message = content.get('message', 'No message provided')

            return {
                'echo': input_message,
                'timestamp': datetime.now().isoformat(),
                'status': 'success'
            }
        except json.JSONDecodeError as e:
            return {
                'error': f'Invalid JSON input: {str(e)}',
                'status': 'failed'
            }
        except Exception as e:
            return {
                'error': str(e),
                'status': 'failed'
            }

    def handle_reverse_text(self, message):
        """Handle text reversal capability"""
        print("🔍 Processing reverse_text request")

        try:
            content = json.loads(message.content)
            text = content.get('text', '')

            reversed_text = text[::-1]

            return {
                'original': text,
                'reversed': reversed_text,
                'timestamp': datetime.now().isoformat(),
                'status': 'success'
            }
        except json.JSONDecodeError as e:
            return {
                'error': f'Invalid JSON input: {str(e)}',
                'status': 'failed'
            }
        except Exception as e:
            return {
                'error': str(e),
                'status': 'failed'
            }

    def handle_analyze_file(self, message):
        """Handle file analysis capability"""
        print("🔍 Processing file analysis request")

        try:
            content = json.loads(message.content)
            filename = content.get('filename', 'unknown')
            file_data = content.get('file')

            if not file_data:
                return {
                    'error': 'No file data provided',
                    'status': 'failed'
                }

            # Decode base64 file data
            try:
                file_bytes = base64.b64decode(file_data)
            except Exception as e:
                return {
                    'error': f'Invalid base64 file data: {str(e)}',
                    'status': 'failed'
                }

            file_size = len(file_bytes)

            # Basic file analysis
            analysis = {
                'size_bytes': file_size,
                'size_kb': round(file_size / 1024, 2),
                'size_mb': round(file_size / (1024 * 1024), 2)
            }

            # Try to detect file type from filename
            if '.' in filename:
                extension = filename.split('.')[-1].lower()
                analysis['extension'] = extension

                # Basic MIME type detection
                mime_types = {
                    'txt': 'text/plain',
                    'json': 'application/json',
                    'xml': 'application/xml',
                    'html': 'text/html',
                    'css': 'text/css',
                    'js': 'application/javascript',
                    'png': 'image/png',
                    'jpg': 'image/jpeg',
                    'jpeg': 'image/jpeg',
                    'gif': 'image/gif',
                    'pdf': 'application/pdf',
                    'zip': 'application/zip'
                }
                analysis['mime_type'] = mime_types.get(extension, 'application/octet-stream')
            else:
                analysis['mime_type'] = 'application/octet-stream'

            # Additional analysis based on content
            try:
                # Try to decode as text
                text_content = file_bytes.decode('utf-8', errors='ignore')
                if text_content:
                    analysis['is_text'] = True
                    analysis['line_count'] = len(text_content.split('\n'))
                    analysis['char_count'] = len(text_content)
                else:
                    analysis['is_text'] = False
            except:
                analysis['is_text'] = False

            print(f"📁 Analyzed file: {filename} ({file_size} bytes)")

            return {
                'filename': filename,
                'size': file_size,
                'mime_type': analysis.get('mime_type', 'unknown'),
                'analysis': analysis,
                'timestamp': datetime.now().isoformat(),
                'status': 'success'
            }

        except json.JSONDecodeError as e:
            return {
                'error': f'Invalid JSON input: {str(e)}',
                'status': 'failed'
            }
        except Exception as e:
            print(f"Error analyzing file: {e}")
            return {
                'error': str(e),
                'status': 'failed'
            }


def main():
    """Main entry point"""
    worker = ExampleWorker()

    # Handle graceful shutdown
    import signal
    import sys

    def signal_handler(sig, frame):
        print('\n🛑 Received shutdown signal, stopping worker...')
        worker.stop()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        worker.start()
    except Exception as e:
        print(f"Failed to start worker: {e}")
        sys.exit(1)


if __name__ == '__main__':
    main()