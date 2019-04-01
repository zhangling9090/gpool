package gpool

import (
	"errors"
	"sync"
	"time"
)

const defaultCleanUpInterval int = 2
const TaskChanCap int32 = 15000
const TaskChanNum int = 4

var ErrPoolReleased = errors.New("gpool has already released.")

// 协程池
type Pool struct {
	// worker数量上限
	capacity int32
	// Task channels
	tchans chan Task
	// 空闲worker队列
	free []*worker
	// 运行中worker数量
	running int32
	// 保护空闲worker列表
	lock sync.Mutex
	// 释放
	once    sync.Once
	release int32
}

type worker struct {
	// 所属协程池
	p           *Pool
	recycleTime time.Time
	stoped      int32
}

type Task func()
