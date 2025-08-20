module github.com/containers/libhvee

go 1.23.3

toolchain go1.23.9

require (
	github.com/containers/common v0.64.1
	github.com/containers/storage v1.59.1
	github.com/go-ole/go-ole v1.3.0
	github.com/onsi/ginkgo/v2 v2.24.0
	github.com/onsi/gomega v1.38.0
	github.com/schollz/progressbar/v3 v3.18.0
	github.com/sirupsen/logrus v1.9.3
	github.com/ulikunitz/xz v0.5.13
	golang.org/x/sys v0.35.0
)

require (
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/term v0.34.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v1.10.1 // This version is used to publish retraction of v1.10.0
	v1.10.0 // Typo version tag for 0.10.0
)
