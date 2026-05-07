package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Register is the doll's first message: UUID and ephemeral X25519 pubkey.
// Plaintext on the wire because the session key does not exist yet.
type Register struct {
	DollUUID   [16]byte
	DollPubKey [32]byte
}

const registerSize = 16 + 32

func (r *Register) MarshalBinary() ([]byte, error) {
	out := make([]byte, registerSize)
	copy(out[0:16], r.DollUUID[:])
	copy(out[16:48], r.DollPubKey[:])
	return out, nil
}

func (r *Register) UnmarshalBinary(b []byte) error {
	if len(b) < registerSize {
		return fmt.Errorf("protocol: register body %d bytes, need %d", len(b), registerSize)
	}
	copy(r.DollUUID[:], b[0:16])
	copy(r.DollPubKey[:], b[16:48])
	return nil
}

// RegisterAck holds the server's ephemeral pubkey in the clear plus a sealed
// initial config. The doll reads ServerPubKey, derives the session key, then
// opens Sealed.
type RegisterAck struct {
	ServerPubKey [32]byte
	Sealed       []byte // crypto.Seal output: nonce + ciphertext
}

func (a *RegisterAck) MarshalBinary() ([]byte, error) {
	out := make([]byte, 32+len(a.Sealed))
	copy(out[0:32], a.ServerPubKey[:])
	copy(out[32:], a.Sealed)
	return out, nil
}

func (a *RegisterAck) UnmarshalBinary(b []byte) error {
	if len(b) < 32 {
		return fmt.Errorf("protocol: register-ack body %d bytes, need at least 32", len(b))
	}
	copy(a.ServerPubKey[:], b[0:32])
	a.Sealed = append([]byte(nil), b[32:]...)
	return nil
}

// Config is the doll's initial behavior config. Marshaled, sealed, and
// embedded in RegisterAck.
type Config struct {
	SleepMs   uint64 // beacon interval in milliseconds
	JitterPct uint8  // 0-100, percent of SleepMs to randomize by
}

func (c *Config) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 0, binary.MaxVarintLen64+1)
	buf = binary.AppendUvarint(buf, c.SleepMs)
	buf = append(buf, c.JitterPct)
	return buf, nil
}

func (c *Config) UnmarshalBinary(b []byte) error {
	sleep, n := binary.Uvarint(b)
	if n <= 0 {
		return errors.New("protocol: bad sleep_ms varint")
	}
	if len(b) < n+1 {
		return errors.New("protocol: missing jitter_pct")
	}
	c.SleepMs = sleep
	c.JitterPct = b[n]
	return nil
}
