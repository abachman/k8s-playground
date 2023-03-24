BINARY_NAME=echo
SOURCE_FILE=server.go

.PHONY: build clean build-image

build:
	GOARCH=arm64 GOOS=darwin go build -o build/${BINARY_NAME}-darwin ${SOURCE_FILE}
	GOARCH=arm64 GOOS=linux go build -o build/${BINARY_NAME}-linux ${SOURCE_FILE}

clean: build/${BINARY_NAME}-darwin build/${BINARY_NAME}-linux
	go clean
	rm -f build/${BINARY_NAME}-darwin
	rm -f build/${BINARY_NAME}-linux

build-image: build/${BINARY_NAME}-linux
	docker build -t abachman/echo-local .
