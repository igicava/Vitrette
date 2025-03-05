protoc:
	go mod tidy &
	export PATH="$PATH:$(go env GOPATH)/bin" &
	protoc --go_out=. \
		--go-grpc_out=. \
		--grpc-gateway_out=. \
		--grpc-gateway_opt generate_unbound_methods=true \
		--openapiv2_out . \
		./api/order.proto

build:
	go build -o main cmd/main.go

run:
	./main