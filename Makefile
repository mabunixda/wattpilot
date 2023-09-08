all: fmt build

preprocess: fmt
	go generate ./...
	# format generated files 
	go fmt ./... 

fmt:
	go fmt ./...

build:
	go build -o shell ./shell/
