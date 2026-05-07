package protocol

import (
	"bytes"
	"errors"
	"testing"

	"marionette/shared/crypto"
)

func TestFrameRoundTrip(t *testing.T) {
	body := []byte("hello dolls")
	v, mt, got, err := Unframe(Frame(MsgBeacon, body))
	if err != nil {
		t.Fatalf("unframe: %v", err)
	}
	if v != Version || mt != MsgBeacon || !bytes.Equal(got, body) {
		t.Fatalf("got v=%d type=%s body=%q", v, mt, got)
	}
}

func TestUnframeBadVersion(t *testing.T) {
	bad := []byte{Version + 1, byte(MsgBeacon)}
	if _, _, _, err := Unframe(bad); !errors.Is(err, ErrVersion) {
		t.Fatalf("got %v, want ErrVersion", err)
	}
}

func TestPackUnpackRoundTrip(t *testing.T) {
	doll, _ := crypto.GenerateX25519KeyPair()
	server, _ := crypto.GenerateX25519KeyPair()
	dollKey, _ := crypto.DeriveSharedKey(doll.Private, server.Public)
	serverKey, _ := crypto.DeriveSharedKey(server.Private, doll.Public)

	body := []byte{1, 2, 3, 4, 5}
	sealed, err := Pack(dollKey, MsgBeaconAck, body)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}
	v, mt, got, err := Unpack(serverKey, sealed)
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}
	if v != Version || mt != MsgBeaconAck || !bytes.Equal(got, body) {
		t.Fatalf("got v=%d type=%s body=%v", v, mt, got)
	}
}
