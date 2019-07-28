package trace

import (
	"context"
	"time"
)

type infoOfContext struct{}

var (
	keyForInfoOfContext = infoOfContext{}
)

// Info 存储了 trace 的必要信息。
//
//
// 注意：这里并不包含任何跟超时相关的信息，
// 如果要设置超时，应该使用 `context.WithDeadline` 或 `context.WithTimeout`。
type Info struct {
	Traceid  Traceid
	Spanid   Spanid
	CSpanid  CSpanid
	Logid    Logid
	Locale   Locale
	Timezone Timezone

	HintCode    HintCode
	HintContent HintContent

	SrcMethod string
	Caller    string
	Callee    string

	// FakeTime 用户自定义时间，在一些仿真的环境下使用
	FakeNow time.Time
}

// Merge 合并 target 到 info，只有 target 里面有值的字段才会被复制过来。
func (info *Info) Merge(target *Info) {
	if target == nil {
		return
	}

	if target.Traceid != "" {
		info.Traceid = target.Traceid
	}

	if target.Spanid != "" {
		info.Spanid = target.Spanid
	}

	if target.CSpanid != "" {
		info.CSpanid = target.CSpanid
	}

	if target.Logid != "" {
		info.Logid = target.Logid
	}

	if target.Locale != "" {
		info.Locale = target.Locale
	}

	if target.Timezone != "" {
		info.Timezone = target.Timezone
	}

	if !target.HintCode.IsEmpty() {
		info.HintCode = target.HintCode
	}

	if target.HintContent != "" {
		info.HintContent = target.HintContent
	}

	if target.SrcMethod != "" {
		info.SrcMethod = target.SrcMethod
	}

	if target.Caller != "" {
		info.Caller = target.Caller
	}

	if target.Callee != "" {
		info.Callee = target.Callee
	}

	if !target.FakeNow.IsZero() {
		info.FakeNow = target.FakeNow
	}
}

// Trace 根据 info 的值来创建来创建 trace 信息。
func (info *Info) Trace() Trace {
	tr := Trace{}

	if info.Traceid != "" {
		tr.setTraceid(info.Traceid)
	}

	if info.CSpanid != "" {
		tr.setCSpanid(info.CSpanid)
	}

	if info.Spanid != "" {
		tr.setSpanid(info.Spanid)
	}

	if info.Logid != "" {
		tr.setLogid(info.Logid)
	}

	if info.Locale != "" {
		tr.setLocale(info.Locale)
	}

	if info.Timezone != "" {
		tr.setTimezone(info.Timezone)
	}

	if !info.HintCode.IsEmpty() {
		tr.setHintCode(info.HintCode)
	}

	if info.HintContent != "" {
		tr.setHintContent(info.HintContent)
	}

	if info.SrcMethod != "" {
		tr.setSrcMethod(info.SrcMethod)
	}

	if info.Caller != "" {
		tr.setCaller(info.Caller)
	}

	if info.Callee != "" {
		tr.setCallee(info.Callee)
	}

	if !info.FakeNow.IsZero() {
		tr.setFakeNow(info.FakeNow)
	}

	return tr
}

// WithInfo 将自定义的 trace 信息加入到 ctx 里面。
func WithInfo(ctx context.Context, info *Info) context.Context {
	if prevInfo, ok := ctx.Value(keyForInfoOfContext).(*Info); ok {
		vInfo := *prevInfo
		vInfo.Merge(info)
		info = &vInfo
	}

	return context.WithValue(ctx, keyForInfoOfContext, info)
}

func parseInfo(ctx context.Context) *Info {
	info, ok := ctx.Value(keyForInfoOfContext).(*Info)

	if !ok {
		return nil
	}

	return info
}
