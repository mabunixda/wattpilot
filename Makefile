all: fmt build

preprocess:
	go generate ./...

fmt:
	go fmt ./...

build:
	go build -o ./shell ./shell/
