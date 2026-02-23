// Package util provides DNS-safe Base32 encoding and generic chunking utilities.
package util

import (
	"encoding/base32"
	"strings"
)

// DNSAlphabet is a Base32 alphabet safe for DNS labels:
// lowercase a-z + digits 2-7, no padding character.
// All resolvers accept [a-z0-9], lowercase avoids case-folding issues.
const DNSAlphabet = "abcdefghijklmnopqrstuvwxyz234567"

// dnsEncoding is the base32 encoding using the DNS-safe alphabet, no padding.
var dnsEncoding = base32.NewEncoding(DNSAlphabet).WithPadding(base32.NoPadding)

// EncodeToLabel encodes raw bytes to a DNS-safe Base32 string.
func EncodeToLabel(data []byte) string {
	return dnsEncoding.EncodeToString(data)
}

// DecodeFromLabel decodes a DNS-safe Base32 string to raw bytes.
// Input is lowercased first to tolerate case-insensitive DNS resolvers.
func DecodeFromLabel(s string) ([]byte, error) {
	return dnsEncoding.DecodeString(strings.ToLower(s))
}
