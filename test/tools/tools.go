//go:build tools
// +build tools

package tools

// Importing the packages here will allow to vendor those via
// `go mod vendor`.

import (
	_ "github.com/onsi/ginkgo/v2/ginkgo"
)
