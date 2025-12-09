.PHONY: proto run build clean run-client docker-up docker-down build-java-worker

proto:
	protoc --go_out=. --go-grpc_out=. --plugin=protoc-gen-go=$(HOME)/go/bin/protoc-gen-go --plugin=protoc-gen-go-grpc=$(HOME)/go/bin/protoc-gen-go-grpc proto/hub.proto

run:
	go run cmd/hub/main.go

run-client:
	go run cmd/client/main.go

build:
	go build -o bin/hub cmd/hub/main.go

build-java-worker:
	cd services/java-maven-worker-v1 && mvn clean package

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

clean:
	rm -rf bin/
	rm -f internal/proto/*.pb.go