bin-deps:  ### Install binary dependencies
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2
.PHONY: bin-deps

migrate-create:  ### Create new migration
	migrate create -ext sql -dir migrations 'pvz'
.PHONY: migrate-create

integration-test: ### Run integration tests
	docker compose -f docker-compose.test.yaml up --abort-on-container-exit --exit-code-from integration --build
.PHONY: integration-test

generate:  ### Run go generate
	go generate ./...
.PHONY: generate

generate-proto: ### Run protoc
	protoc --proto_path=api --go_out=paths=source_relative:internal/controller/grpc/pvz_v1 --go-grpc_out=paths=source_relative:internal/controller/grpc/pvz_v1 api/pvz.proto
.PHONY: generate-proto
