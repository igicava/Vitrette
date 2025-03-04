protoc:
	go mod tidy &
	export PATH="$PATH:$(go env GOPATH)/bin" &
	protoc --go_out=./pkg/api/test \
		--go_opt=paths=source_relative \
		--go-grpc_out=./pkg/api/test \
		--go-grpc_opt=paths=source_relative \
		./api/*.proto

build:
	go build -o main cmd/main.go

run:
	./main