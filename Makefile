export GOOS=windows
export GOARCH=amd64
SRC = $(shell find . -type f -name '*.go')

.PHONY: default
default: build

GOLANGCI_LINT_VERSION := 1.62.2
bin/golangci-lint:
	VERSION=$(GOLANGCI_LINT_VERSION) ./hack/install_golangci.sh

.PHONY: .install.golangci-lint
.install.golangci-lint: bin bin/golangci-lint

.PHONY: validate
validate: bin/golangci-lint
	./bin/golangci-lint run  --skip-dirs "test/e2e"

.PHONY: build
build: validate bin bin/kvpctl.exe bin/dumpvms.exe bin/createvm.exe bin/updatevm.exe

bin:
	mkdir -p bin

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor
	go mod verify


bin/kvpctl.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/kvpctl

bin/dumpvms.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/dumpvms

bin/createvm.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/createvm

bin/updatevm.exe: $(SRC) go.mod go.sum
	go build -o bin ./cmd/updatevm

clean:
	rm -rf bin
