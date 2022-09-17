all: preprocess fmt build

preprocess:
	go generate ./...

fmt:
	go fmt ./...

build:
	CURPWD=$(PWD)
	cd shell
	go build
	cd $(CURPWD)
