package gpool

import (
	"sync/atomic"
)

func newWorker(p *Pool) *worker {
	w := &worker{
		p:    p,
		task: make(chan Task),
	}
	w.run()
	return w
}

func (w *worker) run() {
	go func() {
		for t := range w.task {
			t.Run()
			w.putback()
		}
	}()
}

func (w *worker) putback() {
	w.p.lock.Lock()
	w.p.free = append(w.p.free, w)
	w.p.lock.Unlock()
	w.p.fsig <- struct{}{}
}

func (w *worker) pushtask(t Task) {
	w.task <- t
}

func (w *worker) stop() {
	close(w.task)
	atomic.AddInt32(&w.p.running, int32(-1))
}
