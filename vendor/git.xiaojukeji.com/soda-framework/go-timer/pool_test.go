package timer

import (
	"math/rand"
	"testing"
	"time"
)

type timerInfo struct {
	timer           *Timer
	done            <-chan struct{}
	start           time.Time
	end             time.Time
	expectedTimeout time.Duration
	actualTimeout   time.Duration
}

func TestPoolAfter(t *testing.T) {
	const size = 1 << 16
	var info [size]timerInfo
	init := 200 * time.Millisecond
	max := 300 * time.Millisecond
	p := Precision()
	now := time.Now()

	for i := 0; i < size; i++ {
		timeout := init + time.Duration(i)*max/size + time.Duration(rand.Intn(int(max/size)))
		cur := time.Now()
		timer := After(timeout - cur.Sub(now))
		info[i] = timerInfo{
			timer:           timer,
			start:           now,
			done:            timer.Done(),
			expectedTimeout: timeout,
		}
	}

	cnt := 0
	defaultPool.mu.Lock()

	for i := 0; i < len(defaultPool.timerChans); i++ {
		tc := defaultPool.timerChans[i]

		for tc != nil {
			cnt++
			tc = tc.next
		}
	}

	defaultPool.mu.Unlock()

	for i := 0; i < size; i++ {
		select {
		case <-info[i].done:
			info[i].end = time.Now()
			info[i].actualTimeout = info[i].end.Sub(info[i].start)

			if !info[i].timer.Expired() {
				t.Fatalf("done chan is closed but timer is not expired. [idx:%v]", i)
			}
		}
	}

	if maxTimerCnt := int((max + init) / p); cnt > maxTimerCnt {
		t.Fatalf("there should be less timer. [max:%v] [actual:%v]", maxTimerCnt, cnt)
	}

	for i := 0; i < size; i++ {
		if info[i].expectedTimeout-3*p/4 > info[i].actualTimeout || info[i].expectedTimeout+p < info[i].actualTimeout {
			t.Fatalf("timeout is not accurate enough. [expected:%v] [actual:%v] [idx:%v]", info[i].expectedTimeout, info[i].actualTimeout, i)
		}
	}
}

func TestNewRandomTimers(t *testing.T) {
	const timeoutRange = int64(time.Second)
	const totalCnt = 1 << 20
	var timers [totalCnt]timerInfo
	p := MinPrecision
	pool := NewPool(p)
	now := time.Now()

	for i := 0; i < totalCnt; i++ {
		if i%2 == 1 {
			timers[i].timer = pool.After(time.Duration(rand.Int63n(timeoutRange)))
		} else {
			timers[i].timer = pool.New(now.Add(time.Duration(rand.Int63n(timeoutRange))))
		}
	}

	for i := 0; i < totalCnt; i++ {
		if rand.Int63n(totalCnt) < totalCnt/2 {
			timers[i].done = timers[i].timer.Done()
		} else {
			timers[i].done = closedChan
		}
	}

	cnt := 0
	pool.mu.Lock()

	for i := 0; i < len(pool.timerChans); i++ {
		tc := pool.timerChans[i]

		for tc != nil {
			cnt++
			tc = tc.next
		}
	}

	pool.mu.Unlock()

	if precision := pool.Precision(); precision != p {
		t.Fatalf("precision doesn't match. [expected:%v] [actual:%v]", p, precision)
	}

	if maxTimerCnt := int(timeoutRange / int64(p)); cnt > maxTimerCnt {
		t.Fatalf("there should be less timer. [max:%v] [actual:%v]", maxTimerCnt, cnt)
	}

	for i := 0; i < totalCnt; i++ {
		select {
		case <-timers[i].done:
		}
	}
}

// 用来比较 Timer 和 time.Timer 的性能差异。
func TestTimer(t *testing.T) {
	const timeoutRange = int64(time.Second)
	const totalCnt = 1 << 20
	var doneChans [totalCnt]<-chan time.Time

	for i := 0; i < totalCnt; i++ {
		doneChans[i] = time.After(time.Duration(rand.Int63n(timeoutRange)))
	}

	for i := 0; i < totalCnt; i++ {
		select {
		case <-doneChans[i]:
		}
	}
}
