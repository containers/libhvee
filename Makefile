export GOOS=windows
export GOARCH=amd64
SRC = $(shell find . -type f -name '*.go')

GOLANGCI_LINT_VERSION := 1.54.2
.PHONY: .install.golangci-lint
.install.golangci-lint:
	VERSION=$(GOLANGCI_LINT_VERSION) ./hack/install_golangci.sh

.PHONY: validate
validate:
	./bin/golangci-lint run

.PHONY: default
default: build

.PHONY: build 
build: bin bin/kvpctl.exe bin/dumpvms.exe bin/wmigen bin/createvm.exe bin/dumpsummary.exe bin/updatevm.exe

bin:
	mkdir -p bin

bin/kvpctl.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/kvpctl

bin/dumpvms.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/dumpvms

bin/createvm.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/createvm

bin/updatevm.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/updatevm

bin/dumpsummary.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/dumpsummary

bin/wmigen: export GOOS=
bin/wmigen: export GOARCH=
bin/wmigen: $(SRC) go.mod go.sum
	go build -o bin ./cmd/wmigen

clean:
	rm -rf bin
