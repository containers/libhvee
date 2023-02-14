package kvp

import (
	"errors"
)

var (
	// ErrUnableToWriteToKVP is used when we are unable to write to the kernel
	// device for hyperv
	ErrUnableToWriteToKVP = errors.New("failed to write to hv_kvp")
	// ErrUnableToReadFromKVP is used when we are unable to read from the kernel
	// device for hyperv
	ErrUnableToReadFromKVP = errors.New("failed to read from hv_kvp")
	// ErrNoKeyValuePairsFound means we were unable to find key-value pairs as passed
	// from the hyperv host to this guest.
	ErrNoKeyValuePairsFound = errors.New("unable to find kvp keys")
)

const (
	// Timeout amount of time in ms to poll the hyperv kernel device
	Timeout                   = 1000
	KvpOpRegister1            = 100
	HvSOk                     = 0
	HvKvpExchangeMaxValueSize = 2048
	HvKvpExchangeMaxKeySize   = 512
	KvpOpSet                  = 1
	// KvpKernelDevice s the hyperv kernel device used for communicating key-values pairs
	// on hyperv between the host and guest
	KvpKernelDevice = "/dev/vmbus/hv_kvp"
	// DefaultKVPPoolID is where Windows host write to for Linux VMs
	DefaultKVPPoolID = 0
)

type hvKvpExchgMsgValue struct {
	valueType uint32
	keySize   uint32
	valueSize uint32
	key       [HvKvpExchangeMaxKeySize]uint8
	value     [HvKvpExchangeMaxValueSize]uint8
}

type hvKvpMsgSet struct {
	data hvKvpExchgMsgValue
}

type hvKvpHdr struct {
	operation uint8
	pool      uint8
	pad       uint16
}

type hvKvpMsg struct {
	kvpHdr hvKvpHdr
	kvpSet hvKvpMsgSet
	// unused is needed to get to the same struct size as the C version.
	unused [4856]byte
}

type hvKvpMsgRet struct {
	error  int
	kvpSet hvKvpMsgSet
	// unused is needed to get to the same struct size as the C version.
	unused [4856]byte
}

type PoolID uint8

type ValuePair struct {
	Value string
	Pool  PoolID
}
