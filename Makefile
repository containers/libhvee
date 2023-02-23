export GOOS=windows
export GOARCH=amd64
SRC = $(shell find . -type f -name '*.go')

.PHONY: default
default: build

.PHONY: build 
build: bin bin/kvpctl.exe bin/dumpvms.exe bin/wmigen bin/createvm.exe

bin:
	mkdir -p bin

bin/kvpctl.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/kvpctl

bin/dumpvms.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/dumpvms

bin/createvm.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/createvm

bin/wmigen: export GOOS=
bin/wmigen: export GOARCH=
bin/wmigen: $(SRC) go.mod go.sum
	go build -o bin ./cmd/wmigen

clean:
	rm -rf bin
