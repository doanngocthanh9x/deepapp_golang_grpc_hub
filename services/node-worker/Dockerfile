# Dockerfile for Node Worker
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY services/node-worker/package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY services/node-worker/ ./

# Runtime stage (same as builder for Node)
FROM node:20-alpine

WORKDIR /app

COPY --from=builder /app /app

ENV NODE_ENV=production
ENV HUB_ADDRESS=localhost:50051
ENV WORKER_ID=node-worker

ENTRYPOINT ["node", "worker-plugin-system.js"]
