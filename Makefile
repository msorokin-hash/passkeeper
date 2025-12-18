.PHONY: all build build-client build-server gen-certs clean-certs \
        gen-proto gen-proto-mocks clean-proto test clean-cache \
		fmt help clean check-deps

CERT_DNS = localhost
CERT_IP = 0.0.0.0
CERT_DAYS = 360
CLI_BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
CLI_BUILD_VERSION = 1.0.0

all: build-client build-server

build: all

help:
	@echo "Available targets:"
	@echo "  all          	 - Build both client and server"
	@echo "  build-client 	 - Build client application"
	@echo "  build-server 	 - Build server application"
	@echo "  gen-certs    	 - Generate SSL certificates"
	@echo "  clean-certs  	 - Remove SSL certificates"
	@echo "  gen-proto    	 - Generate protobuf files"
	@echo "  gen-proto-mocks - Generate protobuf mocks"
	@echo "  clean-proto  	 - Remove generated protobuf files"
	@echo "  test         	 - Run tests with coverage"
	@echo "  clean-cache  	 - Clean test cache"
	@echo "  fmt          	 - Format Go code"
	@echo "  clean        	 - Clean all generated files"
	@echo "  help         	 - Show this help"

check-deps:
	@which protoc > /dev/null || (echo "Error: protoc not found. Install protobuf compiler" && exit 1)
	@which openssl > /dev/null || (echo "Error: openssl not found" && exit 1)

build-client:
	@go build -ldflags "-X 'main.Version=$(CLI_BUILD_VERSION)' -X 'main.BuildTime=$(CLI_BUILD_DATE)'" \
		-o cmd/client/client cmd/client/main.go

build-server:
	@go build -o cmd/server/server cmd/server/main.go

gen-certs: check-deps
	@rm -rf ssl
	@mkdir -p ssl
	@echo "subjectAltName=DNS:$(CERT_DNS),IP:$(CERT_IP)" > ssl/server-ext.cnf
	@openssl req -x509 -newkey rsa:4096 -days 365 -nodes \
		-keyout ssl/ca-key.pem -out ssl/ca-cert.pem \
		-subj "/C=RU/ST=Moscow/O=Learn/OU=Education/CN=*"
	@openssl x509 -in ssl/ca-cert.pem -noout -text
	@openssl req -newkey rsa:4096 -nodes \
		-keyout ssl/server-key.pem -out ssl/server-req.pem \
		-subj "/C=RU/ST=Moscow/L=Moscow/O=Learn/OU=Education/CN=*"
	@openssl x509 -req -in ssl/server-req.pem -days $(CERT_DAYS) \
		-CA ssl/ca-cert.pem -CAkey ssl/ca-key.pem -CAcreateserial \
		-out ssl/server-cert.pem -extfile ssl/server-ext.cnf
	@openssl x509 -in ssl/server-cert.pem -noout -text

clean-certs:
	@rm -rf ssl

gen-proto: check-deps
	@rm -f internal/protobuf/*.go
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/protobuf/user.proto
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		internal/protobuf/vault.proto

gen-proto-mocks:
	@mockgen -source=internal/protobuf/user_grpc.pb.go -destination=internal/protobuf/mocks/mock_user_grpc.gen.go -package=mocks
	@mockgen -source=internal/protobuf/vault_grpc.pb.go -destination=internal/protobuf/mocks/mock_vault_grpc.gen.go -package=mocks

clean-proto:
	@rm -f internal/protobuf/*.go

test:
	@go test ./... -cover

clean-cache:
	@go clean -testcache

fmt:
	@goimports -w .
	@gofmt -s -w .

clean: clean-certs clean-proto
	@rm -f cmd/client/client cmd/server/server