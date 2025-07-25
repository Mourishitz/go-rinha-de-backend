up-rinha:
	@echo "Running up-rinha..."
	@docker compose -f ./docker/rinha/docker-compose.yml up

up-d-rinha:
	@echo "Running up-rinha on detached mode..."
	@docker compose -f ./docker/rinha/docker-compose.yml up -d
	@echo "Rinha is running in detached mode."

down-rinha:
	@echo "Stopping up-rinha..."
	@docker compose -f ./docker/rinha/docker-compose.yml down
	@echo "Rinha stopped."

build-proxy:
	@echo "Cleaning previous builds..."
	@rm -f ./docker/go/builds/go-proxy/goProxy
	@echo "Building proxy..."
	@cd ./go-proxy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../docker/go/builds/go-proxy/goProxy ./cmd/
	@cd ..
	@echo "Proxy built successfully."

build-doctor:
	@echo "Cleaning previous builds..."
	@rm -f ./docker/go/builds/go-proxy/goDoctor
	@echo "Building doctor..."
	@cd ./go-doctor && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../docker/go/builds/go-doctor/goDoctor ./cmd
	@cd ..
	@echo "Doctor built successfully."

build-worker:
	@echo "Cleaning previous builds..."
	@rm -f ./docker/go/builds/go-worker/goWorker
	@echo "Building worker..."
	@cd ./go-worker && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ../docker/go/builds/go-worker/goWorker ./cmd
	@cd ..
	@echo "Worker built successfully."
