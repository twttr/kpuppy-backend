.PHONY: build run test clean dev

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

dev:
	go run ./cmd/server

test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

deps:
	go mod download
	go mod tidy

lint:
	golangci-lint run

docker-build:
	docker build -t kpuppy-backend .

docker-run:
	docker run -p 8080:8080 -v $(PWD)/data:/data -e DB_PATH=/data/kpuppy.db kpuppy-backend
