package uuid

import (
	"bytes"
)

var nullByteString = []byte("null")

// MarshalText returns the string-representation of the UUID as a byte-array.
func (u UUID) MarshalText() ([]byte, error) {
	/* Inlined UUID.String() implementation, cannot reuse the one from
	   UUID.String() as that cast will force an additional memory
	   allocation because current version the compiler (go 1.3.1) cannot
	   realize that the data pointer of the result of the UUID.String() call
	   can be moved to the byte-slice return. Speed difference from inlining
	   is about +20% */
	/* Equivalent code:
	   return []byte(u.String()), nil */
	/* NOTE: If this is changed, also look at the following methods:
	   UUID.MarshalJSON() and UUID.String() */
	b := [36]byte{}

	for i, n := range []int{
		0, 2, 4, 6,
		9, 11,
		14, 16,
		19, 21,
		24, 26, 28, 30, 32, 34,
	} {
		b[n] = halfbyte2hexchar[(u[i]>>4)&0x0f]
		b[n+1] = halfbyte2hexchar[u[i]&0x0f]
	}

	b[8] = '-'
	b[13] = '-'
	b[18] = '-'
	b[23] = '-'

	return b[:], nil
}

// MarshalJSON returns the string-representation of the UUID as a JSON-string.
func (u UUID) MarshalJSON() ([]byte, error) {
	/* Needs a slightly different code yet inlined to prevent extra memory
	   allocations, it needs to be 38 bytes and the indices are shifted once
	   to the right to make room for quotation marks */
	/* Equivalent code:
	   return []byte("\"" + u.String() + "\""), nil */
	/* NOTE: If this is changed, also look at the following methods:
	   UUID.MarshalText() and UUID.String() */
	b := [38]byte{}

	for i, n := range []int{
		1, 3, 5, 7,
		10, 12,
		15, 17,
		20, 22,
		25, 27, 29, 31, 33, 35,
	} {
		b[n] = halfbyte2hexchar[(u[i]>>4)&0x0f]
		b[n+1] = halfbyte2hexchar[u[i]&0x0f]
	}

	b[0] = '"'
	b[9] = '-'
	b[14] = '-'
	b[19] = '-'
	b[24] = '-'
	b[37] = '"'

	return b[:], nil
}

// UnmarshalText reads an UUID from a string into the UUID instance.
// If this fails the state of the UUID is undetermined.
func (u *UUID) UnmarshalText(data []byte) error {
	return u.ReadBytes(data)
}

// UnmarshalJSON reads an UUID from a JSON-string into the UUID instance.
// If this fails the state of the UUID is undetermined.
func (u *UUID) UnmarshalJSON(data []byte) error {
	return u.ReadBytes(data[1 : len(data)-1])
}

// MarshalJSON marshals a potentially null UUID into either a string-
// representation of the UUID or the null-constant depending on the
// Valid property.
func (n NullUUID) MarshalJSON() ([]byte, error) {
	if n.Valid {
		return n.UUID.MarshalJSON()
	} else {
		return nullByteString, nil
	}
}

// MarshalText marshals a potentially null UUID into either a string-
// representation of the UUID or the null-constant depending on the
// Valid property.
func (n NullUUID) MarshalText() ([]byte, error) {
	if n.Valid {
		return n.UUID.MarshalText()
	} else {
		return nullByteString, nil
	}
}

// UnmarshalJSON parses a potentially null UUID into a NullUUI instance.
// If the source is the null-constant ("null"), Valid is set to false,
// otherwise it will attempt to parse the given string into the UUID property
// of the NullUUID instance, setting Valid to true if no error is encountered.
// If an error is encountered, Valid is set to false.
func (n *NullUUID) UnmarshalJSON(data []byte) error {
	if bytes.Compare(data, nullByteString) == 0 {
		n.Valid = false
		return nil
	}

	dataLen := len(data)
	if dataLen < 2 {
		n.Valid = false
		return nil
	}

	err := n.UUID.ReadBytes(data[1 : len(data)-1])

	/* Because we modify in place we set Valid to false in case of an error */
	n.Valid = err == nil

	return err
}

// UnmarshalText parses a potentially null UUID into a NullUUI instance.
// If the source is the null-constant ("null"), Valid is set to false,
// otherwise it will attempt to parse the given string into the UUID property
// of the NullUUID instance, setting Valid to true if no error is encountered.
// If an error is encountered, Valid is set to false.
func (n *NullUUID) UnmarshalText(data []byte) error {
	if bytes.Compare(data, nullByteString) == 0 {
		n.Valid = false

		return nil
	}

	err := n.UUID.ReadBytes(data)

	/* Because we modify in place we set Valid to false in case of an error */
	n.Valid = err == nil

	return err
}
