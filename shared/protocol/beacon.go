package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Status is the protocol-level outcome of a String, returned in a Result.
// The Result body's structure is type-specific.
type Status uint8

const (
	StatusOK      Status = 0
	StatusError   Status = 1
	StatusTimeout Status = 2
)

func (s Status) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusError:
		return "Error"
	case StatusTimeout:
		return "Timeout"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// Result is one entry in a Beacon: the outcome of a previously-issued String.
type Result struct {
	StringID uint64
	Status   Status
	Body     []byte
}

// String is one entry in a BeaconAck: a tasking issued from server to doll.
// The Type field selects which payload encoder applies. See strings.go for
// the type catalog and any per-type payload helpers.
type String struct {
	StringID uint64
	Type     StringType
	Payload  []byte
}

// itemHeaderSize is the fixed-width prefix for a Result or String:
// 8-byte string_id + 1-byte status/type, before the varint length.
const itemHeaderSize = 8 + 1

func appendResult(buf []byte, r *Result) []byte {
	buf = binary.LittleEndian.AppendUint64(buf, r.StringID)
	buf = append(buf, byte(r.Status))
	buf = binary.AppendUvarint(buf, uint64(len(r.Body)))
	return append(buf, r.Body...)
}

func readResult(b []byte, r *Result) (int, error) {
	if len(b) < itemHeaderSize {
		return 0, fmt.Errorf("protocol: result header %d bytes, need %d", len(b), itemHeaderSize)
	}
	r.StringID = binary.LittleEndian.Uint64(b[0:8])
	r.Status = Status(b[8])
	n, read := binary.Uvarint(b[itemHeaderSize:])
	if read <= 0 {
		return 0, errors.New("protocol: bad result body length varint")
	}
	bodyStart := itemHeaderSize + read
	bodyEnd := bodyStart + int(n)
	if bodyEnd > len(b) {
		return 0, fmt.Errorf("protocol: result body of %d bytes truncated", n)
	}
	r.Body = append([]byte(nil), b[bodyStart:bodyEnd]...)
	return bodyEnd, nil
}

func appendString(buf []byte, s *String) []byte {
	buf = binary.LittleEndian.AppendUint64(buf, s.StringID)
	buf = append(buf, byte(s.Type))
	buf = binary.AppendUvarint(buf, uint64(len(s.Payload)))
	return append(buf, s.Payload...)
}

func readString(b []byte, s *String) (int, error) {
	if len(b) < itemHeaderSize {
		return 0, fmt.Errorf("protocol: string header %d bytes, need %d", len(b), itemHeaderSize)
	}
	s.StringID = binary.LittleEndian.Uint64(b[0:8])
	s.Type = StringType(b[8])
	n, read := binary.Uvarint(b[itemHeaderSize:])
	if read <= 0 {
		return 0, errors.New("protocol: bad string payload length varint")
	}
	payloadStart := itemHeaderSize + read
	payloadEnd := payloadStart + int(n)
	if payloadEnd > len(b) {
		return 0, fmt.Errorf("protocol: string payload of %d bytes truncated", n)
	}
	s.Payload = append([]byte(nil), b[payloadStart:payloadEnd]...)
	return payloadEnd, nil
}

// Beacon is the doll's pull. Body carries any Results the doll has ready
// to deliver from previously-issued Strings. Empty Results is fine: the
// beacon itself is the "I'm alive, anything for me?" signal.
type Beacon struct {
	Results []Result
}

func (b *Beacon) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, binary.MaxVarintLen64)
	out = binary.AppendUvarint(out, uint64(len(b.Results)))
	for i := range b.Results {
		out = appendResult(out, &b.Results[i])
	}
	return out, nil
}

func (b *Beacon) UnmarshalBinary(buf []byte) error {
	n, read := binary.Uvarint(buf)
	if read <= 0 {
		return errors.New("protocol: bad result count varint")
	}
	offset := read
	b.Results = make([]Result, n)
	for i := range b.Results {
		consumed, err := readResult(buf[offset:], &b.Results[i])
		if err != nil {
			return fmt.Errorf("protocol: result %d: %w", i, err)
		}
		offset += consumed
	}
	return nil
}

// BeaconAck is the server's reply: any pending Strings, possibly empty.
type BeaconAck struct {
	Strings []String
}

func (a *BeaconAck) MarshalBinary() ([]byte, error) {
	out := make([]byte, 0, binary.MaxVarintLen64)
	out = binary.AppendUvarint(out, uint64(len(a.Strings)))
	for i := range a.Strings {
		out = appendString(out, &a.Strings[i])
	}
	return out, nil
}

func (a *BeaconAck) UnmarshalBinary(buf []byte) error {
	n, read := binary.Uvarint(buf)
	if read <= 0 {
		return errors.New("protocol: bad string count varint")
	}
	offset := read
	a.Strings = make([]String, n)
	for i := range a.Strings {
		consumed, err := readString(buf[offset:], &a.Strings[i])
		if err != nil {
			return fmt.Errorf("protocol: string %d: %w", i, err)
		}
		offset += consumed
	}
	return nil
}
