package trace

import (
	"net"
	"testing"
	"time"
)

func TestTraceid(t *testing.T) {
	traceSequence = 0x123456
	localIP = net.IPv4(10, 0, 1, 127).To4()
	processID = 0xfade

	SetTraceidSource(0x54)

	now := int64(0x12345678 * time.Second)
	traceid := MakeTraceid(now)

	if expected := Traceid("0a00017f123456780000fade12345754"); expected != traceid {
		t.Fatalf("invalid traceid. [expected:%v] [actual:%v]", expected, traceid)
	}

	if !traceid.IsValid() {
		t.Fatalf("traceid should be valid. [traceid:%v]", traceid)
	}

	if !localIP.Equal(traceid.IP()) {
		t.Fatalf("traceid's IP is invalid. [expected:%v] [actual:%v]", localIP, traceid.IP())
	}

	if expected := traceSequence; int(expected) != traceid.SequenceID() {
		t.Fatalf("traceid's SequenceID is invalid. [expected:%v] [actual:%v]", expected, traceid.SequenceID())
	}

	if expected := processID; int(expected) != traceid.ProcessID() {
		t.Fatalf("traceid's ProcessID is invalid. [expected:%v] [actual:%v]", expected, traceid.ProcessID())
	}

	if expected := now; int(expected/int64(time.Second)) != traceid.Timestamp() {
		t.Fatalf("traceid's Timestamp is invalid. [expected:%v] [actual:%v]", expected, traceid.Timestamp())
	}

	if expected := traceSource; expected != traceid.Source() {
		t.Fatalf("traceid's Source is invalid. [expected:%v] [actual:%v]", expected, traceid.Source())
	}
}

func BenchmarkNewTraceid(b *testing.B) {
	unixnano := time.Now().UnixNano()

	for i := 0; i < b.N; i++ {
		MakeTraceid(unixnano)
	}
}
