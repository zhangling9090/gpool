package gpool

import (
	"sync"
)

// 协程池
/*
协程池初始化启动一些worker,此时均为空闲状态
*/
type sig struct{}

type Pool struct {
	// worker数量上限
	capacity int32
	// 空闲worker队列
	free []*worker
	// 有新的空闲worker的信号通知channel
	fsig chan sig
	// 运行中worker数量
	running int32
	// 保护空闲worker列表
	lock sync.Mutex
}

type worker struct {
	// 所属协程池
	p *Pool
	// 待执行任务channel
	task chan Task
}

type Task interface {
	Run()
}
