package autoapprover

import "time"

type retryTimer struct {
	start         time.Time
	baseInc       int
	timeEdge      time.Duration
	retryInterval int
}

func newRetryTimer(retry int, retryMax int) *retryTimer {
	return &retryTimer{
		start:         time.Now(),
		baseInc:       retry,
		timeEdge:      time.Duration(retryMax) * time.Second,
		retryInterval: retry,
	}
}

func (t *retryTimer) isTimeOut() bool {
	return time.Since(t.start) >= t.timeEdge
}

func (t *retryTimer) retry() {
	time.Sleep(time.Duration(t.baseInc) * time.Second)
	t.baseInc += t.retryInterval
}
