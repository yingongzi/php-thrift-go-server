package trace

import (
	"testing"
)

func TestHintCode(t *testing.T) {
	if hc := MakeHintCode(""); !hc.IsEmpty() || hc.IsUltron() {
		t.Fatalf("invalid hint code. [hc:%#v]", hc)
	}

	if hc := MakeHintCode("0"); !hc.IsEmpty() || hc.IsUltron() {
		t.Fatalf("invalid hint code. [hc:%#v]", hc)
	}

	if hc := MakeHintCode("1"); hc.IsEmpty() || !hc.IsUltron() {
		t.Fatalf("invalid hint code. [hc:%#v]", hc)
	}

	if hc := MakeHintCode("2"); hc.IsEmpty() || hc.IsUltron() {
		t.Fatalf("invalid hint code. [hc:%#v]", hc)
	}

	if hc := MakeHintCode("2a"); !hc.IsEmpty() || hc.IsUltron() {
		t.Fatalf("invalid hint code. [hc:%#v]", hc)
	}

	if hc := MakeUltronHintCode(); hc != ultronHintCodeMask {
		t.Fatalf("unexpected ultron hint code. [expected:%v] [actual:%v]", ultronHintCodeMask, hc)
	}
}
