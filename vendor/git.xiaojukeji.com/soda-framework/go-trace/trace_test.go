package trace

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

func TestTrace(t *testing.T) {
	traceid := "abcdef"
	spanid := "123456"
	cspanid := "654321"
	hintCode := "234567"
	hintContent := "{\"app_timeout_ms\":20000,\"Cityid\":1,\"lang\":\"zh-CN\",\"utc_offset\":\"480\"}"
	locale := "pt-BR"
	timezone := "America/Fortaleza"
	now := time.Now()
	fakeTime := time.Now().Add(-24 * time.Hour)
	timeout := 10 * time.Millisecond
	time.Sleep(timeout)

	ctx := NewContext(context.Background(), Trace{
		keyTraceid:     traceid,
		keySpanid:      spanid,
		keyCSpanid:     cspanid,
		keyHintCode:    hintCode,
		keyHintContent: hintContent,
		keyTimeout:     fmt.Sprint(int64(timeout)),
		keyNow:         fmt.Sprint(now.UnixNano()),
		keyLocale:      locale,
		keyTimezone:    timezone,
		keyFakeNow:     fmt.Sprint(fakeTime.UnixNano()),
	})
	tr := FromContext(ctx)

	str := "traceid=abcdef||spanid=654321||locale=pt-BR||timezone=America/Fortaleza||hintCode=234567"
	if actual := tr.String(); str != actual {
		t.Fatalf("string should be equal. [expected:%v] [actual:%v]", str, actual)
	}

	if tr.Traceid().String() != traceid {
		t.Fatalf("traceid should be derived. [expected:%v] [actual:%v]", traceid, tr.Traceid())
	}

	if tr.HintCode().String() != hintCode {
		t.Fatalf("hintCode should be derived. [expected:%v] [actual:%v]", hintCode, tr.HintCode())
	}

	if tr.HintContent().String() != hintContent {
		t.Fatalf("hintContent should be derived. [expected:%v] [actual:%v]", hintContent, tr.HintContent())
	}

	if tr.Spanid().String() != cspanid {
		t.Fatalf("spanid should be derived from cspanid. [expected:%v] [actual:%v]", cspanid, tr.Spanid())
	}

	if tr.Locale().String() != locale {
		t.Fatalf("locale should be derived. [expected:%v] [actual:%v]", locale, tr.Locale())
	}

	if tr.Timezone().String() != timezone {
		t.Fatalf("v should be derived. [expected:%v] [actual:%v]", timezone, tr.Timezone())
	}

	if actual := tr.Timeout(); actual > timeout {
		t.Fatalf("timeout should be derived. [expected:%v] [actual:%v]", timeout, actual)
	}
	if fakeTime.IsZero() {
		if n := tr.Now(); n.Equal(now) || n.Before(now) {
			t.Fatalf("now should not be available. [prev:%v] [cur:%v]", now, n)
		} else if deadline, ok := tr.deadline(); !ok {
			t.Fatalf("deadline is not available.")
		} else if expected := n.Add(tr.timeout() - tr.ElapsedTime()); deadline != expected {
			t.Fatalf("deadline is invalid. [expected:%v] [actual:%v]", expected, deadline)
		}
	} else {
		if n := tr.Now(); !n.Equal(fakeTime) {
			t.Fatalf("now should not be available. [prev:%v] [cur:%v]", fakeTime, n)
		} else if deadline, ok := tr.deadline(); !ok {
			t.Fatalf("deadline is not available.")
		} else if expected := n.Add(tr.timeout() - tr.ElapsedTime()); deadline != expected {
			t.Fatalf("deadline is invalid. [expected:%v] [actual:%v]", expected, deadline)
		}
	}

	if deadline, ok := ctx.Deadline(); !ok || deadline.Add(-timeout).After(time.Now()) {
		t.Fatalf("deadline is too large or invalid. [deadline:%v]", deadline)
	}

	select {
	case <-ctx.Done():
	case <-time.After(timeout * 2):
		t.Fatalf("fail to wait for ctx timeout.")
	}

	if err := ctx.Err(); err != context.DeadlineExceeded {
		t.Fatalf("ctx timeout err is invalid. [expected:%v] [actual:%v]", context.DeadlineExceeded, err)
	}
	time.Sleep(timeout)

	ctx = NewContext(context.Background(), tr)
	tr = FromContext(ctx)

	if actual := tr.Now(); !actual.Equal(fakeTime) {
		t.Fatalf("fail to set now. [expected:%v] [actual:%v]", fakeTime, actual)
	}
	except := time.Now().Sub(now)
	if actual := tr.ElapsedTime(); math.Abs(float64(except-actual)/1e6) > 1 {
		t.Fatalf("Elapse should be derived. [expected:%v] [actual:%v]", except, actual)
	}

	if actual := tr.Timeout(); actual > 0 {
		t.Fatalf("timeout should be derived. [expected:%v] [actual:%v]", timeout, actual)
	}
}

func TestEmptyTrace(t *testing.T) {
	tr := Trace{}

	if delta := tr.Now().Sub(time.Now()); delta > 0 || delta < -10*time.Microsecond {
		t.Fatalf("trace now should be the same as time.Now(). [delta:%v]", delta)
	}

	if timeout := tr.Timeout(); timeout != maxTimeout {
		t.Fatalf("trace timeout should be an invalid value. [timeout:%v]", timeout)
	}

	if elapsed := tr.ElapsedTime(); elapsed != 0 {
		t.Fatalf("trace elapsed should be 0. [elapsed:%v]", elapsed)
	}
}

func TestTraceGetterSetters(t *testing.T) {
	if tr := FromContextMayEmpty(context.Background()); tr != nil {
		t.Fatalf("tr should be nil in an empty context. [tr:%v]", tr)
	}

	ctx := NewContext(context.Background(), nil)

	if tr := FromContextMayEmpty(ctx); tr == nil {
		t.Fatalf("tr should not be nil in a ctx created by NewContext.")
	}

	tr := FromContext(context.Background())

	if tr == nil {
		t.Fatalf("tr should be not nil as normalized.")
	}

	if deadline, ok := tr.deadline(); ok || !deadline.IsZero() {
		t.Fatalf("tr should not have deadline. [ok:%v] [deadline:%v]", ok, deadline)
	}

	traceid := MakeTraceid(0)
	tr.setTraceid(traceid)

	if actual := tr.Traceid(); actual != traceid {
		t.Fatalf("fail to set traceid. [expected:%v] [actual:%v]", traceid, actual)
	}

	spanid := makeSpanid(0, 0)
	tr.setSpanid(spanid)

	if actual := tr.Spanid(); !spanid.IsValid() || actual != spanid {
		t.Fatalf("fail to set spanid. [expected:%v] [actual:%v]", spanid, actual)
	}

	cspanid := MakeCSpanid(0)
	tr.setCSpanid(cspanid)

	if actual := tr.CSpanid(); !cspanid.IsValid() || actual != cspanid {
		t.Fatalf("fail to set cspanid. [expected:%v] [actual:%v]", cspanid, actual)
	}

	logid := MakeLogid(0)
	tr.setLogid(logid)

	if actual := tr.Logid(); actual != logid {
		t.Fatalf("fail to set logid. [expected:%v] [actual:%v]", logid, actual)
	}

	hintCode := MakeHintCode("13579")
	tr.setHintCode(hintCode)

	if actual := tr.HintCode(); actual != hintCode {
		t.Fatalf("fail to set hintCode. [expected:%v] [actual:%v]", hintCode, actual)
	}

	hintContent := MakeHintContent("{\"app_timeout_ms\":20000,\"Cityid\":1,\"lang\":\"zh-CN\",\"utc_offset\":\"480\"}")
	tr.setHintContent(hintContent)

	if actual := tr.HintContent(); actual != hintContent {
		t.Fatalf("fail to set hintContent. [expected:%v] [actual:%v]", hintContent, actual)
	}

	srcMethod := "foo"
	tr.setSrcMethod(srcMethod)

	if actual := tr.SrcMethod(); actual != srcMethod {
		t.Fatalf("fail to set srcMethod. [expected:%v] [actual:%v]", srcMethod, actual)
	}

	caller := "bar"
	tr.setCaller(caller)

	if actual := tr.Caller(); actual != caller {
		t.Fatalf("fail to set caller. [expected:%v] [actual:%v]", caller, actual)
	}

	callee := "player"
	tr.setCallee(callee)

	if actual := tr.Callee(); actual != callee {
		t.Fatalf("fail to set callee. [expected:%v] [actual:%v]", callee, actual)
	}

	if expected := fmt.Sprintf("traceid=%v||spanid=%v||hintCode=%v", traceid, spanid, hintCode); expected != tr.String() {
		t.Fatalf("invalid tr.String(). [expected:%v] [actual:%v]", expected, tr)
	}

	tr.setHintCode(0)
	if expected := fmt.Sprintf("traceid=%v||spanid=%v", traceid, spanid); expected != tr.String() {
		t.Fatalf("invalid tr.String(). [expected:%v] [actual:%v]", expected, tr)
	}

	now := time.Now()
	tr.setNow(now)
	if actual := tr.Now(); !actual.Equal(now) {
		t.Fatalf("fail to set now. [expected:%v] [actual:%v]", now, actual)
	}

	fakeTime := time.Now().Add(-time.Hour * 24)
	tr.setFakeNow(fakeTime)

	if actual := tr.Now(); !actual.Equal(fakeTime) {
		t.Fatalf("fail to set now. [expected:%v] [actual:%v]", fakeTime, actual)
	}
	ctx = NewContext(context.Background(), Trace{
		keyCaller:   "a-caller",
		keyCallee:   "a-callee",
		keyHintCode: "12345",
	})
	tr = FromContext(ctx)

	if expected, actual := fmt.Sprintf("traceid=%v||spanid=%v||hintCode=%v", tr.Traceid(), tr.Spanid(), tr.HintCode()), ContextString(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextString(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "a-caller", ContextCaller(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextCaller(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "context-caller", ContextCaller(WithInfo(ctx, &Info{Caller: "context-caller"})); expected != actual {
		t.Fatalf("invalid trace.ContextCaller(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "a-callee", ContextCallee(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextCallee(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "context-callee", ContextCallee(WithInfo(ctx, &Info{Callee: "context-callee"})); expected != actual {
		t.Fatalf("invalid trace.ContextCallee(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := tr.HintCode(), ContextHintCode(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextHintCode(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := HintCode(123), ContextHintCode(WithInfo(ctx, &Info{HintCode: 123})); expected != actual {
		t.Fatalf("invalid trace.ContextHintCode(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	ctx = context.Background()

	if expected, actual := emptyContextString, ContextString(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextString(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "", ContextCaller(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextCaller(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "", ContextCallee(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextCallee(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := emptyHintCode, ContextHintCode(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextHintCode(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	info := &Info{
		Caller: "previous-caller",
		Callee: "previous-callee",
	}
	ctx = NewContext(ctx, info.Trace())

	if expected, actual := "previous-caller", ContextCaller(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextCaller(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "previous-callee", ContextCallee(ctx); expected != actual {
		t.Fatalf("invalid trace.ContextCallee(ctx). [expected:%v] [actual:%v]", expected, actual)
	}

	if expected, actual := "a-caller", ContextCaller(WithInfo(WithInfo(ctx, info), &Info{Caller: "a-caller"})); expected != actual {
		t.Fatalf("invalid trace.ContextCaller(ctx). [expected:%v] [actual:%v]", expected, actual)
	}
}

func TestFromContextMayEmpty(t *testing.T) {
	ctx := context.Background()

	if tr := FromContextMayEmpty(ctx); tr != nil {
		t.Fatalf("ctx should not contain any info. [tr:%v, %#v]", tr, tr)
	}

	ctx = WithInfo(ctx, &Info{
		Traceid: "123456",
	})

	if tr := FromContextMayEmpty(ctx); tr == nil {
		t.Fatalf("ctx should contain some info. [tr:%v, %#v]", tr, tr)
	}
}

func BenchmarkTraceString(b *testing.B) {
	ctx := NewContext(context.Background(), nil)

	for i := 0; i < b.N; i++ {
		tr := FromContext(ctx)
		_ = tr.String()
	}
}

func BenchmarkTraceContextString(b *testing.B) {
	ctx := NewContext(context.Background(), nil)

	for i := 0; i < b.N; i++ {
		ContextString(ctx)
	}
}
