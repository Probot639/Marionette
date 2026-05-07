// Package protocol defines the wire format for doll-to-teamserver traffic.
//
// The frame layout, message catalog, and bootstrapping rules live in
// README.md alongside this file. Read that before adding a new message
// type or changing field encodings.
package protocol

import (
	"errors"
	"fmt"

	"marionette/shared/crypto"
)

// Version is the current protocol version. Bump on any wire-format change
// that is not backwards-compatible.
const Version byte = 1

// nonceSize matches the AES-GCM nonce length used by shared/crypto (RFC 5116).
const nonceSize = 12

// headerSize is the inner header: version + type.
const headerSize = 2

// MessageType identifies the kind of message inside a frame.
type MessageType byte

const (
	MsgRegister    MessageType = 1
	MsgRegisterAck MessageType = 2
	MsgBeacon      MessageType = 3
	MsgBeaconAck   MessageType = 4
)

func (t MessageType) String() string {
	switch t {
	case MsgRegister:
		return "Register"
	case MsgRegisterAck:
		return "RegisterAck"
	case MsgBeacon:
		return "Beacon"
	case MsgBeaconAck:
		return "BeaconAck"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

var (
	ErrShortFrame = errors.New("protocol: frame shorter than header")
	ErrVersion    = errors.New("protocol: unsupported version")
)

// Frame builds an inner frame: [version:u8][type:u8][body...].
// Use directly for plaintext messages (Register, RegisterAck outer).
// Pipe through Pack for everything else.
func Frame(t MessageType, body []byte) []byte {
	out := make([]byte, headerSize+len(body))
	out[0] = Version
	out[1] = byte(t)
	copy(out[2:], body)
	return out
}

// Unframe splits an inner frame into version, type, and body.
// The returned body is a sub-slice of in, not a copy.
func Unframe(in []byte) (byte, MessageType, []byte, error) {
	if len(in) < headerSize {
		return 0, 0, nil, ErrShortFrame
	}
	v, t := in[0], MessageType(in[1])
	if v != Version {
		return v, t, nil, fmt.Errorf("%w: got %d", ErrVersion, v)
	}
	return v, t, in[2:], nil
}

// Pack frames a body and AEAD-seals it with the doll's session key.
// Output: [nonce:12][ciphertext].
func Pack(key []byte, t MessageType, body []byte) ([]byte, error) {
	return crypto.Seal(key, Frame(t, body))
}

// Unpack opens a sealed envelope and unframes the plaintext.
func Unpack(key, sealed []byte) (byte, MessageType, []byte, error) {
	framed, err := crypto.Open(key, sealed)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("open: %w", err)
	}
	return Unframe(framed)
}
