# Dockerfile for Java Worker
FROM maven:3.9-eclipse-temurin-21 AS builder

WORKDIR /app

# Copy pom.xml
COPY services/java-simple-worker/pom.xml ./

# Download dependencies
RUN mvn dependency:go-offline

# Copy source code
COPY services/java-simple-worker/src ./src

# Build
RUN mvn clean package -DskipTests

# Runtime stage
FROM eclipse-temurin:21-jre-alpine

WORKDIR /app

COPY --from=builder /app/target/java-simple-worker-*.jar ./java-worker.jar

ENV HUB_ADDRESS=localhost:50051
ENV WORKER_ID=java-simple-worker

ENTRYPOINT ["java", "-jar", "java-worker.jar"]
