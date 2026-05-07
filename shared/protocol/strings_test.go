package protocol

import (
	"bytes"
	"testing"
)

func TestSleepRoundTrip(t *testing.T) {
	for _, ms := range []uint64{0, 1, 1000, 30000, 1<<32 - 1, 1 << 63} {
		b := MarshalSleep(ms)
		got, err := UnmarshalSleep(b)
		if err != nil {
			t.Errorf("ms=%d: unmarshal: %v", ms, err)
			continue
		}
		if got != ms {
			t.Errorf("ms=%d: got %d", ms, got)
		}
	}
}

func TestUnmarshalSleepTrailing(t *testing.T) {
	b := append(MarshalSleep(45000), 0xff)
	if _, err := UnmarshalSleep(b); err == nil {
		t.Error("expected error for trailing bytes after sleep varint")
	}
}

// TestStringsInBeaconAck packs the four interesting v1 String shapes (typed +
// opaque payload, typed + structured payload, typed + empty payload, typed +
// path payload) into a BeaconAck and confirms each round-trips with the right
// type tag and the right payload bytes.
func TestStringsInBeaconAck(t *testing.T) {
	in := BeaconAck{
		Strings: []String{
			{StringID: 1, Type: StringShell, Payload: []byte("ls -la")},
			{StringID: 2, Type: StringSleep, Payload: MarshalSleep(45000)},
			{StringID: 3, Type: StringExit, Payload: nil},
			{StringID: 4, Type: StringCd, Payload: []byte("/var/log")},
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
	if len(out.Strings) != 4 {
		t.Fatalf("count: got %d", len(out.Strings))
	}

	if out.Strings[0].Type != StringShell || !bytes.Equal(out.Strings[0].Payload, []byte("ls -la")) {
		t.Errorf("strings[0]: got %+v", out.Strings[0])
	}

	if out.Strings[1].Type != StringSleep {
		t.Errorf("strings[1] type: got %s", out.Strings[1].Type)
	}
	ms, err := UnmarshalSleep(out.Strings[1].Payload)
	if err != nil {
		t.Errorf("strings[1] sleep: %v", err)
	}
	if ms != 45000 {
		t.Errorf("strings[1] sleep ms: got %d, want 45000", ms)
	}

	if out.Strings[2].Type != StringExit || len(out.Strings[2].Payload) != 0 {
		t.Errorf("strings[2]: got %+v", out.Strings[2])
	}

	if out.Strings[3].Type != StringCd || !bytes.Equal(out.Strings[3].Payload, []byte("/var/log")) {
		t.Errorf("strings[3]: got %+v", out.Strings[3])
	}
}
