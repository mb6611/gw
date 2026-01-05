.PHONY: build install clean

BINARY=gw
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X main.version=$(VERSION)" -o $(BINARY) ./cmd/gw

install: build
	mv $(BINARY) /usr/local/bin/

clean:
	rm -f $(BINARY)
