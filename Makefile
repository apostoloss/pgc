run:
	go run ./cmd/server/main.go

test:
	go test ./...

docs:
	swag init -g ./cmd/server/main.go -o ./docs
