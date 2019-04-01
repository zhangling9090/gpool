package gpool

import (
	"sync"
	"sync/atomic"
	"time"
)

func NewPool(cap int32) *Pool {
	p := Pool{
		capacity: cap,
	}
	p.cond = sync.NewCond(&p.lock)
	go p.cleanUp(defaultCleanUpInterval)
	return &p
}

func (p *Pool) cleanUp(interval int) {
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	duration := time.Duration(interval) * time.Second
	for range ticker.C {
		currentTime := time.Now()
		p.lock.Lock()
		workers := p.free
		if len(workers) == 0 && p.running == 0 && atomic.LoadInt32(&p.release) == 1 {
			p.lock.Unlock()
			return
		}

		n := -1
		for i, w := range workers {
			if w == nil || currentTime.Sub(w.recycleTime) < duration {
				break
			}
			n = i
			w.stop()
			workers[i] = nil
		}
		if n > -1 {
			if n >= len(workers)-1 {
				p.free = workers[:0]
			} else {
				p.free = workers[n+1:]
			}
		}
		p.lock.Unlock()
	}
}

func (p *Pool) SubmitTask(t Task) error {
	if atomic.LoadInt32(&p.release) == 1 {
		return ErrPoolReleased
	}
	w, err := p.getWorker()
	if err != nil {
		return err
	}
	w.pushtask(t)
	return nil
}

func (p *Pool) getWorker() (*worker, error) {
	var w *worker
	p.lock.Lock()
	fcount := len(p.free) - 1
	// 有空闲worker
	if fcount >= 0 {
		w = p.free[fcount]
		p.free = p.free[0:fcount]
		p.lock.Unlock()
		return w, nil
	}

	// 未达上限，动态扩容
	if p.running < p.capacity {
		p.lock.Unlock()
		w = newWorker(p)
		return w, nil
	}

	// 已达上限，等待空闲worker到来
	for {
		p.cond.Wait()
		if len(p.free) < 1 {
			continue
		}
		break
	}
	fcount = len(p.free) - 1
	w = p.free[fcount]
	p.free = p.free[0:fcount]
	p.lock.Unlock()

	return w, nil
}

func (p *Pool) Release() {
	p.once.Do(func() {
		atomic.StoreInt32(&p.release, int32(1))
		p.lock.Lock()
		for _, w := range p.free {
			w.stop()
		}
		p.free = p.free[:0]
		p.lock.Unlock()
	})
}
