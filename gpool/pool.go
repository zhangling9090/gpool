package gpool

import (
	"sync/atomic"
)

func NewPool(initSize, cap int32) *Pool {
	p := Pool{
		capacity: cap,
		free:     make([]*worker, initSize, cap),
		fsig:     make(chan sig, cap),
	}

	p.lock.Lock()
	for i := 0; i < int(initSize); i++ {
		w := &worker{p: &p, task: make(chan Task)}
		p.free[i] = w
		w.run()
		p.fsig <- sig{}
	}
	p.lock.Unlock()
	atomic.StoreInt32(&p.running, initSize)
	return &p
}

func (p *Pool) SubmitTask(t Task) error {
	w, err := p.getWorker()
	if err != nil {
		return err
	}
	w.pushtask(t)
	return nil
}

func (p *Pool) getWorker() (*worker, error) {
	p.lock.Lock()
	fcount := len(p.free) - 1
	// 有空闲worker
	if fcount >= 0 {
		<-p.fsig
		w := p.free[fcount]
		p.free = p.free[0:fcount]
		p.lock.Unlock()
		return w, nil
	}
	p.lock.Unlock()

	// 未达上限，动态扩容
	if p.running < p.capacity {
		p.resize()
		return p.getWorker()
	}

	// 已达上限，等待空闲worker到来
	<-p.fsig
	p.lock.Lock()
	fcount = len(p.free) - 1
	w := p.free[fcount]
	p.free = p.free[0:fcount]
	p.lock.Unlock()
	return w, nil
}

// 动态扩容,未达上限则扩大2倍
func (p *Pool) resize() {
	var newCap = p.running * int32(2)
	if newCap > p.capacity {
		newCap = p.capacity
	}
	p.lock.Lock()
	for i := p.running; i < newCap; i++ {
		w := newWorker(p)
		p.free = append(p.free, w)
		p.fsig <- sig{}
		atomic.AddInt32(&p.running, int32(1))
	}
	p.lock.Unlock()

}
