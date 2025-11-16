# bin dir
BIN_DIR := bin
# binary name
BINARY_NAME := code-sandbox-mcp-server

# build the project to the bin directory
build-linux:
	@mkdir -p ${BIN_DIR}
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BIN_DIR}/${BINARY_NAME} ./cmd/code-sandbox-mcp

# build the project to the bin directory
build-darwin:
	@mkdir -p ${BIN_DIR}
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ${BIN_DIR}/${BINARY_NAME} ./cmd/code-sandbox-mcp

# build the project to the bin directory
build-windows:
	@mkdir -p ${BIN_DIR}
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${BIN_DIR}/${BINARY_NAME} ./cmd/code-sandbox-mcp

# clean the compiled file
clean:
	rm -rf ${BIN_DIR}

# initialize the docker image of the programming language
init-images:
	docker pull python:latest
	docker pull php:latest
	docker pull golang:latest