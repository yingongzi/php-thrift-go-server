package timer

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// MinPrecision 是支持的最小精度。
	MinPrecision = 500 * time.Microsecond
)

const (
	ringBufferSize = 1 << 10
)

var (
	closedChan = make(chan struct{})
)

func init() {
	close(closedChan)
}

// Pool 是一个用来存储并跟踪所有 timer 的池子，通过降低 timer 的精度来尽量的复用 chan，
// 降低内存碎片和提升性能。
//
// 这里实现的是一个 Hashing Wheel with Ordered Timer Lists，算法详见：
// https://blog.acolyer.org/2015/11/23/hashed-and-hierarchical-timing-wheels/
type Pool struct {
	mu            sync.Mutex
	timerChanPool sync.Pool

	precision  int64
	timerChans [ringBufferSize]*timerChan
	timerCnt   int64
	loopStart  chan struct{}
}

// NewPool 创建一个新的 timer 池。
func NewPool(precision time.Duration) *Pool {
	if precision < MinPrecision {
		panic(fmt.Errorf("minimum precision is %v but setting it to %v", MinPrecision, precision))
	}

	p := &Pool{
		precision: int64(precision),
		timerChanPool: sync.Pool{
			New: newTimerChan,
		},
		loopStart: make(chan struct{}, 1),
	}
	go p.loop()

	return p
}

// New 创建一个新 Timer。
func (p *Pool) New(expire time.Time) *Timer {
	return &Timer{
		pool:     p,
		expireAt: expire.UnixNano(),
	}
}

// After 创建一个 timer，从当前时间算起，timeout 之后超时。
func (p *Pool) After(timeout time.Duration) *Timer {
	t := int64(timeout)
	precision := p.precision

	if t <= precision/4 {
		return &Timer{
			pool:     p,
			expireAt: 0,
			done:     closedChan,
		}
	} else if t < precision {
		t = precision
	}

	now := time.Now()

	return &Timer{
		pool:     p,
		expireAt: now.UnixNano() + t,
	}
}

// Precision 返回当前 p 的最小精度。
func (p *Pool) Precision() time.Duration {
	return time.Duration(p.precision)
}

func (p *Pool) makeDoneChan(expireAt int64) <-chan struct{} {
	precision := p.precision
	now := time.Now()

	if expireAt <= now.UnixNano()+precision/4 {
		return closedChan
	}

	p.mu.Lock()

	remainder := expireAt % precision
	expireAt -= remainder

	if remainder > precision/2 {
		expireAt += precision
	}

	idx := (expireAt / precision) % ringBufferSize
	tc := p.timerChans[idx]
	created := false

	if tc == nil {
		created = true
		tc = p.timerChanPool.Get().(*timerChan)
		tc.tick = expireAt
		tc.done = make(chan struct{})
		tc.next = nil
		p.timerChans[idx] = tc
	} else if tc.tick > expireAt {
		created = true
		ntc := p.timerChanPool.Get().(*timerChan)
		ntc.tick = expireAt
		ntc.done = make(chan struct{})
		ntc.next = tc
		p.timerChans[idx] = ntc

		tc = ntc
	} else if tc.tick < expireAt {
		next := tc.next

		for next != nil && next.tick < expireAt {
			tc = tc.next
			next = tc.next
		}

		if next == nil || next.tick > expireAt {
			created = true
			ntc := p.timerChanPool.Get().(*timerChan)
			ntc.tick = expireAt
			ntc.done = make(chan struct{})
			ntc.next = next
			tc.next = ntc

			tc = ntc
		} else {
			tc = next
		}
	}

	p.mu.Unlock()

	if created && atomic.AddInt64(&p.timerCnt, 1) == 1 {
		p.loopStart <- struct{}{}
	}

	return tc.done
}

func (p *Pool) loop() {
	<-p.loopStart

	prev := time.Now()
	precision := p.precision
	halfPrecision := precision / 2

	for range time.Tick(time.Duration(halfPrecision)) {
		now := time.Now()
		tick := prev.UnixNano()
		nextTick := now.UnixNano()
		cnt := int64(0)

		tick -= tick % precision
		p.mu.Lock()

		for tick < nextTick+halfPrecision/2 {
			idx := (tick / precision) % ringBufferSize
			tc := p.timerChans[idx]

			for tc != nil {
				if tc.tick > tick {
					break
				}

				cnt++
				next := tc.next

				close(tc.done)
				tc.done = nil
				tc.next = nil
				p.timerChanPool.Put(tc)

				tc = next
			}

			p.timerChans[idx] = tc
			tick += precision
		}

		p.mu.Unlock()
		prev = now

		if atomic.AddInt64(&p.timerCnt, -cnt) == 0 {
			<-p.loopStart
		}
	}
}

type timerChan struct {
	tick int64
	next *timerChan
	done chan struct{}
}

func newTimerChan() interface{} {
	return new(timerChan)
}
