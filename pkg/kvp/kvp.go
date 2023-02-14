//go:build linux

package kvp

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

// readKvpData reads all key-value pairs from the hyperv kernel device and creates
// a map representation of them
func readKvpData() (map[string]ValuePair, error) {
	ret := map[string]ValuePair{}
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
			ret[string(key)] = ValuePair{
				Value: string(value),
				Pool:  PoolID(hvMsg.kvpHdr.pool),
			}
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

// GetKeyValuePairs reads the key value pairs from the wmi hyperv kernel device
// and returns them in map form.  the map value is a ValuePair which contains
// the value string and the poolid
func GetKeyValuePairs() (map[string]ValuePair, error) {
	return readKvpData()
}

// GetSplitKeyValues "filters" KVPs looking for split values using a key and pool_id.  Returns the assembled
// split values as a key as well as a new KVP that no longer has the split keys in question
func GetSplitKeyValues(key string, pool PoolID, kvps map[string]ValuePair) (string, map[string]ValuePair, error) {
	if len(kvps) < 1 {
		return "", kvps, ErrNoKeyValuePairsFound
	}

	var (
		parts     []string
		counter   = 0
		leftOvers map[string]ValuePair
	)

	// Being extra diligent here
	b, err := json.Marshal(kvps)
	if err != nil {
		return "", nil, err
	}
	if err := json.Unmarshal(b, &leftOvers); err != nil {
		return "", nil, err
	}
	for {
		wantKey := fmt.Sprintf("%s%d", key, counter)
		val, exists := leftOvers[wantKey]
		if !exists {
			break
		}
		if exists && val.Pool == pool {
			parts = append(parts, val.Value)
			// Pop key/value from map
			delete(leftOvers, wantKey)
		}
		counter++
	}
	if len(parts) < 1 {
		return "", kvps, ErrNoKeyValuePairsFound
	}
	return strings.Join(parts, ""), leftOvers, nil
}
