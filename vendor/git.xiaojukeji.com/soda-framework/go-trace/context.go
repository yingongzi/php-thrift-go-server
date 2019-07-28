package trace

import (
	"context"
	"time"

	"git.xiaojukeji.com/soda-framework/go-timer"
)

type valueOfContext struct{}

var (
	keyForValueOfContext = valueOfContext{}
)

// timeoutContext 符合 context.Context 接口，并且实现了更加高效的超时机制，
// 可以有效的在超高并发连接的情况下减少超时判断带来的性能损耗。
type timeoutContext struct {
	context.Context

	deadline time.Time
	timer    *timer.Timer
}

func newTimeoutContext(ctx context.Context, v *value) *timeoutContext {
	c := &timeoutContext{
		Context: ctx,
	}

	if v != nil {
		c.Context = context.WithValue(ctx, keyForValueOfContext, v)

		if deadline, ok := v.Deadline(); ok {
			parentDeadline, ok := ctx.Deadline()

			if !ok || parentDeadline.After(deadline) {
				c.deadline = deadline
				c.timer = timer.New(deadline)
			}
		}
	}

	return c
}

func (c *timeoutContext) Deadline() (deadline time.Time, ok bool) {
	if c.deadline.IsZero() {
		return c.Context.Deadline()
	}

	return c.deadline, true
}

func (c *timeoutContext) Done() <-chan struct{} {
	if c.timer == nil {
		return c.Context.Done()
	}

	return c.timer.Done()
}

func (c *timeoutContext) Err() error {
	if c.timer != nil && c.timer.Expired() {
		return context.DeadlineExceeded
	}

	return c.Context.Err()
}
