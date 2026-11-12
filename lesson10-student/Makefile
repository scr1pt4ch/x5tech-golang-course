.PHONY: run build test cover fmt vet check swagger

run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test ./...

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

fmt:
	gofmt -w .

vet:
	go vet ./...

check: fmt vet test build

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs
