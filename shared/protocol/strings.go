package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// StringType identifies the kind of tasking inside a String. The doll
// dispatches on this byte to pick a payload handler.
type StringType uint8

const (
	StringShell  StringType = 1
	StringSleep  StringType = 2
	StringExit   StringType = 3
	StringCd     StringType = 4
	StringPwd    StringType = 5
	StringWhoami StringType = 6
)

func (t StringType) String() string {
	switch t {
	case StringShell:
		return "Shell"
	case StringSleep:
		return "Sleep"
	case StringExit:
		return "Exit"
	case StringCd:
		return "Cd"
	case StringPwd:
		return "Pwd"
	case StringWhoami:
		return "Whoami"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

// Payload encoders. Most v1 String types carry either opaque bytes
// (Shell, Cd) or no payload at all (Exit, Pwd, Whoami), so they don't
// need helpers. Sleep is the one that does.

// MarshalSleep encodes a Sleep payload: one varint of the new sleep_ms.
func MarshalSleep(ms uint64) []byte {
	return binary.AppendUvarint(nil, ms)
}

// UnmarshalSleep decodes a Sleep payload back to its sleep_ms value.
// Rejects payloads with trailing bytes after the varint.
func UnmarshalSleep(b []byte) (uint64, error) {
	ms, n := binary.Uvarint(b)
	if n <= 0 {
		return 0, errors.New("protocol: bad sleep_ms varint")
	}
	if n != len(b) {
		return 0, fmt.Errorf("protocol: sleep payload has %d trailing bytes", len(b)-n)
	}
	return ms, nil
}
