package trace

import (
	"context"
	"testing"
	"time"
)

func TestContextWithoutTrace(t *testing.T) {
	ctx := newTimeoutContext(context.Background(), nil)

	if deadline, ok := ctx.Deadline(); ok || !deadline.IsZero() {
		t.Fatalf("deadline should be empty. [ok:%v] [deadline:%v]", ok, deadline)
	}

	select {
	case <-ctx.Done():
		t.Fatalf("done chan should never return.")
	default:
	}

	if e := ctx.Err(); e != nil {
		t.Fatalf("ctx should never return error. [err:%v]", e)
	}

	if v, ok := ctx.Value(keyForValueOfContext).(*value); ok || v != nil {
		t.Fatalf("value in ctx should be nil. [v:%v]", v)
	}
}

func TestContextWithTraceWithoutTimeout(t *testing.T) {
	traceid := MakeTraceid(time.Now().UnixNano())
	ctx := newTimeoutContext(context.Background(), &value{
		Info: Info{
			Traceid: traceid,
		},
	})

	if deadline, ok := ctx.Deadline(); ok || !deadline.IsZero() {
		t.Fatalf("deadline should be empty. [ok:%v] [deadline:%v]", ok, deadline)
	}

	select {
	case <-ctx.Done():
		t.Fatalf("done chan should never return.")
	default:
	}

	if e := ctx.Err(); e != nil {
		t.Fatalf("ctx should never return error. [err:%v]", e)
	}

	if v, ok := ctx.Value(keyForValueOfContext).(*value); !ok || v == nil || v.Traceid != traceid {
		t.Fatalf("trace in ctx should be valid. [v:%v]", v)
	}
}

func TestContextWithTraceAndTimeout(t *testing.T) {
	traceid := MakeTraceid(time.Now().UnixNano())
	timeout := 10 * time.Millisecond
	v := &value{
		Info: Info{
			Traceid: traceid,
		},
		Timeout: timeout,
	}
	ctx := newTimeoutContext(context.Background(), v)
	now := time.Now()

	if deadline, ok := ctx.Deadline(); !ok || deadline.IsZero() || deadline.Sub(now) > timeout {
		t.Fatalf("deadline should be valid. [ok:%v] [deadline:%v]", ok, deadline)
	}

	done := ctx.Done()
	time.Sleep(timeout * 2)

	select {
	case <-done:
	default:
		t.Fatalf("done chan should be closed.")
	}

	if done != ctx.Done() {
		t.Fatalf("done chan should be consistent.")
	}

	if e := ctx.Err(); e != context.DeadlineExceeded {
		t.Fatalf("ctx should return DeadlineExceeded error. [err:%v]", e)
	}
}

func TestContextWithTraceAndTimeoutAndParentCtx(t *testing.T) {
	traceid := MakeTraceid(time.Now().UnixNano())
	timeout := 20 * time.Millisecond
	v := &value{
		Info: Info{
			Traceid: traceid,
		},
		Timeout: timeout,
	}
	key := struct{ key int }{}
	value := "bar"
	parentCtx, cancel := context.WithTimeout(context.Background(), timeout/2)
	parentCtx = context.WithValue(parentCtx, key, value)
	defer cancel()
	ctx := newTimeoutContext(parentCtx, v)
	now := time.Now()

	if deadline, ok := ctx.Deadline(); !ok || deadline.IsZero() || deadline.Sub(now) > timeout/2 {
		t.Fatalf("deadline should be valid. [ok:%v] [deadline:%v]", ok, deadline)
	}

	done := ctx.Done()
	time.Sleep(timeout)

	select {
	case <-done:
	default:
		t.Fatalf("done chan should be closed.")
	}

	if done != ctx.Done() {
		t.Fatalf("done chan should be consistent.")
	}

	if e := ctx.Err(); e != context.DeadlineExceeded {
		t.Fatalf("ctx should return DeadlineExceeded error. [err:%v]", e)
	}

	if v := ctx.Value(key); v == nil || v.(string) != value {
		t.Fatalf("fail to get parent value. [v:%v]", v)
	}
}
