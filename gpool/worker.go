package gpool

import (
	"log"
	"sync/atomic"
	"time"
)

func newWorker(p *Pool) *worker {
	w := &worker{
		p:    p,
		task: make(chan Task, TaskChanCap),
	}
	w.run()
	return w
}

func (w *worker) run() {
	atomic.AddInt32(&w.p.running, int32(1))

	go func() {
		defer func() {
			p := recover()
			if p != nil {
				log.Println(p)
			}
		}()

		for t := range w.task {
			if t == nil {
				atomic.AddInt32(&w.p.running, int32(-1))
				return
			}
			//t.Run()
			t()
			w.putback()
		}

	}()
}

func (w *worker) putback() {
	w.recycleTime = time.Now()
	w.p.lock.Lock()
	w.p.free = append(w.p.free, w)
	w.p.cond.Signal()
	w.p.lock.Unlock()
}

func (w *worker) pushtask(t Task) {
	w.task <- t
}

func (w *worker) stop() {
	w.task <- nil
}
