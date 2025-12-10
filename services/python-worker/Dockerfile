# Dockerfile for Python Worker
FROM python:3.10-alpine

WORKDIR /app

# Install system dependencies
RUN apk add --no-cache gcc musl-dev

# Copy requirements
COPY services/python-worker/requirements.txt ./

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY services/python-worker/ ./

ENV PYTHONUNBUFFERED=1
ENV HUB_ADDRESS=localhost:50051
ENV WORKER_ID=python-worker

ENTRYPOINT ["python", "worker_plugin_system.py"]
