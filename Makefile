
VERSION    ?= $(shell git describe --tags --always --dirty)
LDFLAGS    ?= -X main.Version=$(VERSION) -w -s

build: build.local

build.local:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o build/brickd_exporter-$(shell go env GOOS)-$(shell go env GOARCH)

build.raspi:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "$(LDFLAGS)" -o build/brickd_exporter-linux-arm6

clean:
	rm -rf ./build/
