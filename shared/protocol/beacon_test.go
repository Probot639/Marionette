package protocol

import (
	"bytes"
	"testing"

	"marionette/shared/crypto"
)

func TestBeaconRoundTrip(t *testing.T) {
	in := Beacon{
		Results: []Result{
			{StringID: 1, Status: StatusOK, Body: []byte("first result")},
			{StringID: 2, Status: StatusError, Body: []byte("oops")},
			{StringID: 7, Status: StatusTimeout, Body: nil},
		},
	}
	b, err := in.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Beacon
	if err := out.UnmarshalBinary(b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Results) != len(in.Results) {
		t.Fatalf("count: got %d, want %d", len(out.Results), len(in.Results))
	}
	for i := range in.Results {
		want, got := in.Results[i], out.Results[i]
		if want.StringID != got.StringID || want.Status != got.Status || !bytes.Equal(want.Body, got.Body) {
			t.Errorf("result %d: got %+v, want %+v", i, got, want)
		}
	}
}

func TestBeaconAckRoundTrip(t *testing.T) {
	in := BeaconAck{
		Strings: []String{
			{StringID: 100, Type: 1, Payload: []byte("whoami")},
			{StringID: 101, Type: 4, Payload: []byte("/tmp")},
		},
	}
	b, err := in.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out BeaconAck
	if err := out.UnmarshalBinary(b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Strings) != len(in.Strings) {
		t.Fatalf("count: got %d, want %d", len(out.Strings), len(in.Strings))
	}
	for i := range in.Strings {
		want, got := in.Strings[i], out.Strings[i]
		if want.StringID != got.StringID || want.Type != got.Type || !bytes.Equal(want.Payload, got.Payload) {
			t.Errorf("string %d: got %+v, want %+v", i, got, want)
		}
	}
}

func TestBeaconEmpty(t *testing.T) {
	in := Beacon{}
	b, err := in.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(b) != 1 {
		t.Errorf("empty beacon should be 1 byte (varint 0), got %d", len(b))
	}
	var out Beacon
	if err := out.UnmarshalBinary(b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out.Results) != 0 {
		t.Errorf("expected empty results, got %d", len(out.Results))
	}
}

// TestBeaconPackUnpack runs a Beacon through the full encrypt-decrypt path
// to confirm the marshaled body composes cleanly with frame.go's AEAD wrap.
func TestBeaconPackUnpack(t *testing.T) {
	doll, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("doll keygen: %v", err)
	}
	server, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("server keygen: %v", err)
	}
	key, err := crypto.DeriveSharedKey(doll.Private, server.Public)
	if err != nil {
		t.Fatalf("derive: %v", err)
	}

	bn := Beacon{Results: []Result{{StringID: 42, Status: StatusOK, Body: []byte("done")}}}
	body, err := bn.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal beacon: %v", err)
	}
	sealed, err := Pack(key, MsgBeacon, body)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}

	_, mt, gotBody, err := Unpack(key, sealed)
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}
	if mt != MsgBeacon {
		t.Fatalf("type: got %s, want Beacon", mt)
	}
	var got Beacon
	if err := got.UnmarshalBinary(gotBody); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got.Results) != 1 || got.Results[0].StringID != 42 {
		t.Errorf("got %+v", got)
	}
}
