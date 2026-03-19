.PHONY: generate-api
generate-api:
	go generate ./api
	go mod tidy

.PHONY: build
build: generate-api
	go build -o server cmd/main.go

.PHONY: start
start:
	docker-compose build
	docker-compose up

.PHONY: integration-test
integration-test:
	cd integration && go test ./...
