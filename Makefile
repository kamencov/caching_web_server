.PHONY: run_docker
run_docker:
	docker compose -f docker/docker-compose.yaml up --build -d

.PHONY: stop_docker
stop_docker:
	docker compose -f docker/docker-compose.yaml down -v

.PHONY: cover
cover:
	go test -short -count=1 -coverprofile=coverage.out ./...
	@echo "=== Coverage ==="
	@go tool cover -func=coverage.out | tail -n 1 
	go tool cover -html=coverage.out
	rm coverage.out