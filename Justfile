GOTEST := `command -v gotest 2>/dev/null || echo "$(command -v go) test"`

# Show help information
help:
    @just --list

# Run tests
test:
    {{GOTEST}} -v ./...

# Run go vet
vet:
    go vet ./...

# Run golangci-lint
lint:
    golangci-lint run

# Run all checks
check: vet lint test
    @echo "All checks passed!"

# Build the project
build:
    go build -o bin/ ./...

# Clean build artifacts
clean:
    rm -rf bin/

# Show a detailed coverage report
@cover:
    COVER_FILE="$(mktemp /tmp/cover.XXXXXX.out)" \
    && {{GOTEST}} -coverpkg=./... -covermode=count -coverprofile=$COVER_FILE ./... \
    && go tool cover -func=$COVER_FILE

# Show the total coverage
cover-basic:
    {{GOTEST}} -coverpkg=./... -cover ./...

# Open the cover report in your browser
@cover-html:
    COVER_FILE="$(mktemp /tmp/cover.XXXXXX.out)" \
    && {{GOTEST}} -coverpkg=./... -covermode=count -coverprofile=$COVER_FILE ./... \
    && go tool cover -html=$COVER_FILE
    echo -e "\nOpening report in browser..."



judge:
    golangci-lint run --config golangci.yaml ./...

judge-all:
    golangci-lint run --config golangci-all.yaml ./...

judge-critic:
    golangci-lint run --config golangci-critic.yaml ./...
