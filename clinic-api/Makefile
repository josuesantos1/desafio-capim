.PHONY: run
run:
	@go run ./cmd

.PHONY: build
build:
	@go build -o ./bin/clinic-api ./cmd

.PHONY: test
test:
	@go test -race -count=1 ./...

.PHONY: cover
cover:
	@go test -coverprofile=coverage.out $$(go list ./... | grep -v -e /cmd$$ -e /docs$$ -e /mocks$$)
	@go tool cover -html=coverage.out

.PHONY: lint
lint:
	@golangci-lint run ./...

.PHONY: lint-fix
lint-fix:
	@gofmt -w .
	@golangci-lint run --fix ./...

.PHONY: swagger
swagger:
	@swag init -g cmd/main.go -o docs

.PHONY: docker
docker:
	@docker build -t clinic-api .
