all: fmt build

preprocess:
	go generate ./...
	go fmt ./...

fmt:
	go fmt ./...

build:
	go build -o shell ./shell/
