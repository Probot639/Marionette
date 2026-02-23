package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/curve25519"
)

// X25519KeyPair holds an X25519 private/public key pair.
type X25519KeyPair struct {
	Private [32]byte
	Public  [32]byte
}

// GenerateX25519KeyPair generates a new ephemeral X25519 key pair.
func GenerateX25519KeyPair() (*X25519KeyPair, error) {
	kp := &X25519KeyPair{}
	privBytes := make([]byte, 32)
	if _, err := rand.Read(privBytes); err != nil {
		return nil, fmt.Errorf("generate private key: %w", err)
	}
	// Clamp per RFC 7748 §5
	privBytes[0] &= 248
	privBytes[31] &= 127
	privBytes[31] |= 64
	copy(kp.Private[:], privBytes)

	pub, err := curve25519.X25519(kp.Private[:], curve25519.Basepoint)
	if err != nil {
		return nil, fmt.Errorf("derive public key: %w", err)
	}
	copy(kp.Public[:], pub)
	return kp, nil
}

// DeriveSharedKey performs an X25519 Diffie-Hellman exchange and returns
// a 32-byte AES-256 key derived via SHA-256 of the shared secret.
func DeriveSharedKey(localPriv, remotePub [32]byte) ([]byte, error) {
	shared, err := curve25519.X25519(localPriv[:], remotePub[:])
	if err != nil {
		return nil, fmt.Errorf("X25519: %w", err)
	}
	hash := sha256.Sum256(shared)
	return hash[:], nil
}
