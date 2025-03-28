module github.com/containers/libhvee

go 1.23.0

toolchain go1.23.7

require (
	github.com/containers/common v0.62.3
	github.com/containers/storage v1.57.2
	github.com/go-ole/go-ole v1.3.0
	github.com/onsi/ginkgo/v2 v2.23.3
	github.com/onsi/gomega v1.36.3
	github.com/schollz/progressbar/v3 v3.18.0
	github.com/sirupsen/logrus v1.9.3
	github.com/ulikunitz/xz v0.5.12
	golang.org/x/sys v0.31.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/net v0.37.0 // indirect
	golang.org/x/term v0.30.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v1.10.1 // This version is used to publish retraction of v1.10.0
	v1.10.0 // Typo version tag for 0.10.0
)
