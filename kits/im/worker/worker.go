package worker

import (
	"context"

	"github.com/doublemo/baa/cores/pool/worker"
)

var (
	workers worker.WorkerPool
)

// Config 工人配置文件
type Config struct {
	MaxWorkers int `alias:"maxworkers" default:"50"`
}

// Init 初始化
func Init(c Config) {
	workers = *worker.New(c.MaxWorkers)
}

// Submit 提交任务
func Submit(f func()) {
	workers.Submit(f)
}

// Size returns the maximum number of concurrent workers.
func Size() int {
	return workers.Size()
}

// Stop stops the worker pool and waits for only currently running tasks to
// complete.  Pending tasks that are not currently running are abandoned.
// Tasks must not be submitted to the worker pool after calling stop.
//
// Since creating the worker pool starts at least one goroutine, for the
// dispatcher, Stop() or StopWait() should be called when the worker pool is no
// longer needed.
func Stop() {
	workers.Stop()
}

// StopWait stops the worker pool and waits for all queued tasks tasks to
// complete.  No additional tasks may be submitted, but all pending tasks are
// executed by workers before this function returns.
func StopWait() {
	workers.StopWait()
}

// Stopped returns true if this worker pool has been stopped.
func Stopped() bool {
	return workers.Stopped()
}

// SubmitWait enqueues the given function and waits for it to be executed.
func SubmitWait(f func()) {
	workers.SubmitWait(f)
}

// WaitingQueueSize returns the count of tasks in the waiting queue.
func WaitingQueueSize() int {
	return workers.WaitingQueueSize()
}

// Pause causes all workers to wait on the given Context, thereby making them
// unavailable to run tasks.  Pause returns when all workers are waiting.
// Tasks can continue to be queued to the workerpool, but are not executed
// until the Context is canceled or times out.
//
// Calling Pause when the worker pool is already paused causes Pause to wait
// until all previous pauses are canceled.  This allows a goroutine to take
// control of pausing and unpausing the pool as soon as other goroutines have
// unpaused it.
//
// When the workerpool is stopped, workers are unpaused and queued tasks are
// executed during StopWait.
func Pause(ctx context.Context) {
	workers.Pause(ctx)
}
