deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod download

build:
	go build -o build/integration-testing-cli github.com/di-graph/integration-testing-cli/src/cmd
