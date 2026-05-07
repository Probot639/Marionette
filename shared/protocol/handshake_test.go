package protocol

import (
	"bytes"
	"testing"

	"marionette/shared/crypto"
)

func TestRegisterRoundTrip(t *testing.T) {
	in := Register{}
	for i := range in.DollUUID {
		in.DollUUID[i] = byte(i + 1)
	}
	for i := range in.DollPubKey {
		in.DollPubKey[i] = byte(i + 100)
	}
	b, err := in.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if len(b) != registerSize {
		t.Fatalf("size: got %d, want %d", len(b), registerSize)
	}
	var out Register
	if err := out.UnmarshalBinary(b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if in != out {
		t.Fatal("round-trip mismatch")
	}
}

func TestConfigRoundTrip(t *testing.T) {
	in := Config{SleepMs: 30000, JitterPct: 25}
	b, err := in.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out Config
	if err := out.UnmarshalBinary(b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if in != out {
		t.Fatalf("got %+v, want %+v", out, in)
	}
}

// TestHandshake walks the full Register / RegisterAck flow with crypto,
// confirming both sides derive the same session key and that the doll can
// open the server's sealed initial config.
func TestHandshake(t *testing.T) {
	dollKP, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("doll keygen: %v", err)
	}
	serverKP, err := crypto.GenerateX25519KeyPair()
	if err != nil {
		t.Fatalf("server keygen: %v", err)
	}

	// Doll sends Register.
	reg := Register{DollPubKey: dollKP.Public}
	for i := range reg.DollUUID {
		reg.DollUUID[i] = byte(i)
	}
	regBody, err := reg.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal register: %v", err)
	}
	regWire := Frame(MsgRegister, regBody)

	// Server reads Register, derives session key.
	_, mt, gotBody, err := Unframe(regWire)
	if err != nil {
		t.Fatalf("unframe register: %v", err)
	}
	if mt != MsgRegister {
		t.Fatalf("type: got %s, want Register", mt)
	}
	var gotReg Register
	if err := gotReg.UnmarshalBinary(gotBody); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	serverKey, err := crypto.DeriveSharedKey(serverKP.Private, gotReg.DollPubKey)
	if err != nil {
		t.Fatalf("server derive: %v", err)
	}

	// Server seals config, sends RegisterAck.
	cfg := Config{SleepMs: 30000, JitterPct: 25}
	cfgBytes, err := cfg.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	sealed, err := crypto.Seal(serverKey, cfgBytes)
	if err != nil {
		t.Fatalf("seal config: %v", err)
	}
	ack := RegisterAck{ServerPubKey: serverKP.Public, Sealed: sealed}
	ackBody, err := ack.MarshalBinary()
	if err != nil {
		t.Fatalf("marshal ack: %v", err)
	}
	ackWire := Frame(MsgRegisterAck, ackBody)

	// Doll reads RegisterAck, derives the same session key, opens config.
	_, _, gotAckBody, err := Unframe(ackWire)
	if err != nil {
		t.Fatalf("unframe ack: %v", err)
	}
	var gotAck RegisterAck
	if err := gotAck.UnmarshalBinary(gotAckBody); err != nil {
		t.Fatalf("unmarshal ack: %v", err)
	}
	dollKey, err := crypto.DeriveSharedKey(dollKP.Private, gotAck.ServerPubKey)
	if err != nil {
		t.Fatalf("doll derive: %v", err)
	}
	if !bytes.Equal(dollKey, serverKey) {
		t.Fatal("derived keys differ")
	}
	plain, err := crypto.Open(dollKey, gotAck.Sealed)
	if err != nil {
		t.Fatalf("open config: %v", err)
	}
	var gotCfg Config
	if err := gotCfg.UnmarshalBinary(plain); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if gotCfg != cfg {
		t.Fatalf("config: got %+v, want %+v", gotCfg, cfg)
	}
}
