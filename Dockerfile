FROM golang:alpine

WORKDIR /app

COPY . .

RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
RUN go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go
RUN go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
RUN go get -u github.com/golang/protobuf/protoc-gen-go
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN apk update && \
    apk add --no-cache make protobuf-dev \
    && rm -rf /var/cache/apk/*

RUN make protoc

RUN make build

EXPOSE 8081 50051

CMD ["/app/main"]