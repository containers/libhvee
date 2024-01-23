export GOOS=windows
export GOARCH=amd64
SRC = $(shell find . -type f -name '*.go')
VERSION ?= 0.0.1
IMG ?= quay.io/rhqp/libhvee-e2e:v${VERSION}
CONTAINER_MANAGER ?= podman

.PHONY: default
default: build

GOLANGCI_LINT_VERSION := 1.55.2
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

.PHONY: build-oci-e2e
build-oci-e2e: 
	${CONTAINER_MANAGER} build -t ${IMG} -f oci/e2e/Containerfile --build-arg=OS=${GOOS} --build-arg=ARCH=${GOARCH} .
