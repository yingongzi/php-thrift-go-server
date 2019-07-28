package trace

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestContextWithInfo(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	fakeNow := now.Add(-time.Hour * 24)
	expected := &Info{
		Traceid:     MakeTraceid(0),
		Spanid:      makeSpanid(0, 0),
		Logid:       MakeLogid(0),
		HintCode:    MakeHintCode("1234"),
		HintContent: MakeHintContent("{\"app_timeout_ms\":20000,\"Cityid\":1,\"lang\":\"zh-CN\",\"utc_offset\":\"480\"}"),
		SrcMethod:   "test",
		Caller:      "caller",
		Callee:      "callee",
		FakeNow:     fakeNow,
	}
	ctx = WithInfo(ctx, expected)
	actual := parseInfo(ctx)

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("fail to set info into context. [expected:%v] [actual:%v]", expected, actual)
	}
}

func TestInfoMerge(t *testing.T) {
	now := time.Now()
	fakeNow := now.Add(-time.Hour * 24)
	expected := &Info{
		Traceid:     MakeTraceid(0),
		Spanid:      makeSpanid(0, 0),
		Logid:       MakeLogid(0),
		HintCode:    MakeHintCode("2345"),
		HintContent: MakeHintContent("{\"app_timeout_ms\":30000,\"Cityid\":47,\"lang\":\"zh-CN\",\"utc_offset\":\"480\"}"),
		SrcMethod:   "test",
		Caller:      "caller",
		Callee:      "callee",
		FakeNow:     fakeNow,
	}
	actual := &Info{}
	actual.Merge(expected)

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("fail to merge info. [expected:%v] [actual:%v]", expected, actual)
	}
}
