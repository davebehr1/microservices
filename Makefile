temporal-local-up:
	docker-compose up -d

temporal-local-down:
	docker-compose down

check: build lint test-unit

test-unit:
	go test -v -short -race ./...

build:
	go build ./...

lint:
	golangci-lint run

tools:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.34.1

run-square:
	go run square-service/main.go

run-volume:
	go run voluem-service/main.go

run-workflow:
	go run workflow-service/main.go
