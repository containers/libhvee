package ignition

import "bytes"

// Dice takes input and splits it into an array for byte
// arrays.  Each part must be less than the maximum size
// of a kvp value
func Dice(k *bytes.Reader) (Segments, error) {
	var (
		// done is a simple bool indicator that we no longer
		// need to iterate
		done  bool
		parts Segments
	)
	for {
		sl := make([]byte, kvpValueMaxLen)
		n, err := k.Read(sl)
		if err != nil {
			return nil, err
		}
		// if we read and the length is less that the max read,
		// then we are at the end
		if n < kvpValueMaxLen {
			sl = sl[0:n]
			done = true
		}
		parts = append(parts, sl)
		if done {
			break
		}
	}
	return parts, nil
}

// Glue takes an array of byte arrays which represent the values read
// from the kvp device and combines them into one byte array
func Glue(parts Segments) (b []byte) {
	// bytes.Join would be nice here, but it requires a separator
	// byte and inserts the separator even if it is nil or empty
	for _, p := range parts {
		b = append(b, p...)
	}
	return
}
