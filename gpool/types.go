package gpool

import (
	"errors"
	"sync"
	"time"
)

const defaultCleanUpInterval int = 2
const TaskChanCap int32 = 1

var ErrPoolReleased = errors.New("gpool has already been released")

// 协程池
type Pool struct {
	// worker数量上限
	capacity int32
	// 空闲worker队列
	free []*worker
	// 运行中worker数量
	running int32
	// 保护空闲worker列表
	lock sync.Mutex
	// 等待空闲锁
	cond *sync.Cond
	// 释放
	once    sync.Once
	release int32
}

type worker struct {
	// 所属协程池
	p *Pool
	// 待执行任务channel
	task        chan Task
	recycleTime time.Time
}

/*
type Task interface {
	Run()
}
*/
type Task func()
