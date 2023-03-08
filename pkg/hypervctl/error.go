//go:build windows
// +build windows

package hypervctl

import "errors"

// VM State errors
var (
	ErrMachineAlreadyRunning = errors.New("machine already running")
	ErrMachineNotRunning     = errors.New("machine not running")
	ErrMachineStateInvalid   = errors.New("machine in invalid state for action")
	ErrMachineStarting       = errors.New("machine is currently starting")
)

// VM Creation errors
var (
	ErrMachineAlreadyExists = errors.New("machine already exists")
)

type DestroySystemResult int32

// VM Destroy Exit Codes
const (
	VMDestroyCompletedwithNoError DestroySystemResult = 0
	VMDestroyNotSupported         DestroySystemResult = 1
	VMDestroyFailed               DestroySystemResult = 2
	VMDestroyTimeout              DestroySystemResult = 3
	VMDestroyInvalidParameter     DestroySystemResult = 4
	VMDestroyInvalidState         DestroySystemResult = 5
)

func (e DestroySystemResult) Reason() string {
	switch e {
	case VMDestroyNotSupported:
		return "not supported"
	case VMDestroyFailed:
		return "failed"
	case VMDestroyTimeout:
		return "timeout"
	case VMDestroyInvalidParameter:
		return "invalid parameter"
	case VMDestroyInvalidState:
		return "invalid state"
	}
	return "Unknown"
}
