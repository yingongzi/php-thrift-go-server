// Package timer 实现了一个高性能、GC 友好的 timer，不过代价是牺牲了必要的精度。
package timer

import (
	"time"
)

var (
	defaultPool = NewPool(time.Millisecond)
)

// Timer 是一个定时器，可以通过这个结构的各种方法来判断是否超时。
type Timer struct {
	pool     *Pool
	expireAt int64
	done     <-chan struct{}
}

// Done 返回一个 chan，用于等待 t 超时。
// 这个 chan 不返回任何数据，仅通过是否关闭来表达超时，所以只能用于
// for..range 或者 select。
//
// 处于性能考虑，仅在第一次调用 Done 函数的时候才会真正创建 chan，
// 如果没有必要同步等待超时，建议不要调用这个函数。
func (t *Timer) Done() <-chan struct{} {
	if t.done != nil {
		return t.done
	}

	t.done = t.pool.makeDoneChan(t.expireAt)
	return t.done
}

// Expired 返回当前这个 t 是否已经超时。
func (t *Timer) Expired() bool {
	if t.expireAt == 0 {
		return true
	}

	if t.done != nil {
		select {
		case <-t.done:
			return true
		default:
			return false
		}
	}

	now := time.Now()
	return now.UnixNano() >= t.expireAt+int64(t.pool.Precision())
}

// Deadline 返回 t 的超时时间。
func (t *Timer) Deadline() time.Time {
	return time.Unix(t.expireAt/int64(time.Second), t.expireAt%int64(time.Second))
}

// New 创建一个在指定时间超时的 Timer。
func New(expire time.Time) *Timer {
	return defaultPool.After(expire.Sub(time.Now()))
}

// After 创建一个在 timeout 超时的 Timer。
func After(timeout time.Duration) *Timer {
	return defaultPool.After(timeout)
}

// Precision 返回默认 timer 池的最小精度。
func Precision() time.Duration {
	return time.Duration(defaultPool.precision)
}
