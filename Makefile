lint:
	golangci-lint run
	yarn prettier --write .

deps:
	go mod download
	yarn
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.50.0

test:
	go test -count=1 -cover ./...
