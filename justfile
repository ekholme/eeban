default: run

run:
    go run ./cmd/eeban

dev:
    EEBAN_DB=./eeban.dev.db go run ./cmd/eeban

test:
    go test ./...

build:
    CGO_ENABLED=0 go build -o bin/eeban ./cmd/eeban

tidy:
    go mod tidy
