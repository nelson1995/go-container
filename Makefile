PACKAGE_PATH := ./cmd/go-container
BINARY_NAME := go-container

## build: build the application
build:
	go build -o bin/${BINARY_NAME} ${PACKAGE_PATH}

## run: run the application
run:
	sudo ./bin/${BINARY_NAME} run ${CONTAINER_SHELL}