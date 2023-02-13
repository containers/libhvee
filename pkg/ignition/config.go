package ignition

import "errors"

const (
	// kvpValueMaxLen is the maximum real-world length of bytes that can
	// be stored in the value of the wmi key-value pair data exchange
	kvpValueMaxLen = int(990)
)

// segment is a portion of an ignition file represented in a byte array
type segment []byte

// Segments is an array of byte arrays that together make up the entirety of
// an ignition file.
type Segments []segment

var (
	// ErrUnableToWriteToKVP is used when we are unable to write to the kernel
	// device for hyperv
	ErrUnableToWriteToKVP = errors.New("failed to write to hv_kvp")
	// ErrUnableToReadFromKVP is used when we are unable to read from the kernel
	// device for hyperv
	ErrUnableToReadFromKVP = errors.New("failed to read from hv_kvp")
	// ErrNoIgnitionKeysFound means we were unable to find key-value pairs as passed
	// from the hyperv host to this guest.
	ErrNoIgnitionKeysFound = errors.New("unable to find ignition keys")
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
)

var (
	// Key represents the prefix key name for finding ignition file parts
	// in the key value pairs.  it normally will have an integer added to the
	// end when looking up keys sequentially
	Key = "com_coreos_ignition_kvp_"
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
