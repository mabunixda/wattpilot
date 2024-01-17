
IGNORED:=$(shell bash -c "source .metadata.sh ; env | sed 's/=/:=/;s/^/export /' > .metadata.make")

ifeq ($(VERSION),)
	include .metadata.make
else
	# Preserve the passed-in version & iteration (homebrew).
	_VERSION:=$(VERSION)
	_ITERATION:=$(ITERATION)
	include .metadata.make
	VERSION:=$(_VERSION)
	ITERATION:=$(_ITERATION)
endif

all: shell

build: $(BINARY)
$(BINARY): main.go
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(BINARY) -ldflags "-w -s $(VERSION_LDFLAGS)"

exe: $(BINARY).amd64.exe
windows: $(BINARY).amd64.exe
$(BINARY).amd64.exe: main.go
	# Building windows 64-bit x86 binary.
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags "-w -s $(VERSION_LDFLAGS)"

clean:
	rm -f $(BINARY) $(BINARY).amd64.exe
