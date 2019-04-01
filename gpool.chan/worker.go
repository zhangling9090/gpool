package gpool

import (
	"log"
	"sync/atomic"
	"time"
)

func newWorker(p *Pool) *worker {
	w := &worker{
		p: p,
	}
	w.run()
	return w
}

func (w *worker) run() {
	atomic.AddInt32(&w.p.running, int32(1))
	atomic.StoreInt32(&w.stoped, int32(0))
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				log.Println(err)
			}
		}()

		var t Task
		for atomic.LoadInt32(&w.stoped) == int32(0) {
			select {
			case t = <-w.p.tchans:
				t()
			default:
				w.stop()
			}
		}

	}()
}

func (w *worker) putback() {
	w.recycleTime = time.Now()
	w.p.lock.Lock()
	w.p.free = append(w.p.free, w)
	w.p.lock.Unlock()
}

func (w *worker) stop() {
	atomic.AddInt32(&w.p.running, int32(-1))
	atomic.StoreInt32(&w.stoped, int32(1))
	w.putback()
}
