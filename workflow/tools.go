package workflow

import "sync/atomic"

var taskCount int64

func GetTaskCount() int64 {
	return atomic.LoadInt64(&taskCount)
}

func changeTaskCount(delta int64) {
	atomic.AddInt64(&taskCount, delta)
}
