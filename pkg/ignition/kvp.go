//go:build linux

package ignition

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

// readKvpData reads all key-value pairs from the hyperv kernel device and creates
// a map representation of them
func readKvpData() (map[string]string, error) {
	ret := make(map[string]string)

	kvp, err := unix.Open(KvpKernelDevice, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(kvp)

	var (
		hvMsg    hvKvpMsg
		hvMsgRet hvKvpMsgRet
	)

	const sizeOf = int(unsafe.Sizeof(hvMsg))

	var (
		asByteSlice    []byte = (*(*[sizeOf]byte)(unsafe.Pointer(&hvMsg)))[:]
		retAsByteSlice []byte = (*(*[sizeOf]byte)(unsafe.Pointer(&hvMsgRet)))[:]
	)

	hvMsg.kvpHdr.operation = KvpOpRegister1

	l, err := unix.Write(kvp, asByteSlice)
	if err != nil {
		return nil, err
	}
	if l != int(sizeOf) {
		return nil, ErrUnableToWriteToKVP
	}

next:
	for {
		var pfd unix.PollFd
		pfd.Fd = int32(kvp)
		pfd.Events = unix.POLLIN
		pfd.Revents = 0

		howMany, err := unix.Poll([]unix.PollFd{pfd}, Timeout)
		if err != nil {
			if err == unix.EINVAL {
				return nil, err
			} else {
				continue
			}
		}

		if howMany == 0 {
			return ret, nil
		}

		l, err := unix.Read(kvp, asByteSlice)
		if l != sizeOf {
			return nil, ErrUnableToReadFromKVP
		}
		if err != nil {
			return nil, err
		}

		switch hvMsg.kvpHdr.operation {
		case KvpOpRegister1:
			continue next
		case KvpOpSet:
			// on the next two variables, we are cutting the last byte because otherwise
			// it is padded and key lookups fail
			key := []byte(hvMsg.kvpSet.data.key[:hvMsg.kvpSet.data.keySize-1])
			value := []byte(hvMsg.kvpSet.data.value[:hvMsg.kvpSet.data.valueSize-1])
			ret[string(key)] = string(value)
		}

		hvMsgRet.error = HvSOk

		l, err = unix.Write(kvp, retAsByteSlice)
		if err != nil {
			return nil, err
		}
		if l != int(sizeOf) {
			return nil, ErrUnableToWriteToKVP
		}
	}
}

func getIgnitionFromKVP() ([]byte, error) {
	ret, err := readKvpData()
	if err != nil {
		return nil, err
	}
	if len(ret) == 0 {
		return []byte{}, nil
	}
	var (
		counter int
		parts   Segments
	)
	for {
		lookForKey := fmt.Sprintf("%s%d", Key, counter)
		val, exists := ret[lookForKey]
		if !exists {
			break
		}
		parts = append(parts, []byte(val))
		counter++
	}
	if len(parts) < 1 {
		return nil, ErrNoIgnitionKeysFound
	}
	return Glue(parts), nil

}
