//go:build windows
// +build windows

package wmiext

import (
	"github.com/drtimf/wmi"
)

// Variant of Enum.NextObject that also sets object and class paths if present
func NextObjectWithPath(enum *wmi.Enum, target interface{}) (bool, error) {
	var err error

	var instance *wmi.Instance
	if instance, err = enum.Next(); err != nil {
		return false, err
	}

	if instance == nil {
		return true, nil
	}

	defer instance.Close()

	return false, InstanceGetAll(instance, target)
}
