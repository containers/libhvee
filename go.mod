module github.com/containers/libhvee

go 1.25.0

require (
	github.com/go-ole/go-ole v1.3.0
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/schollz/progressbar/v3 v3.19.1
	github.com/sirupsen/logrus v1.9.4
	github.com/ulikunitz/xz v0.5.15
	go.podman.io/common v0.0.0-20250901164813-7046ad001ce8
	go.podman.io/storage v1.59.1-0.20250820085751-a13b38f45723
	golang.org/x/sys v0.46.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.36.0 // indirect
	golang.org/x/net v0.56.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/term v0.44.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
)

retract (
	v1.10.1 // This version is used to publish retraction of v1.10.0
	v1.10.0 // Typo version tag for 0.10.0
)
