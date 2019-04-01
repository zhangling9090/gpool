# gpool
>目标：实现一个支持HttpServer的协程池

## 第一步：让协程池正常工作 ##
>第一版gpool：参考ants开源项目架构
>第二版gpool.chan：gpool维护一个Task Channel，提交任务进channel，worker循环从channel拿任务，拿不到任务就把自己加入到空闲队列，设计初衷是减少worker频繁出队入队的加锁开销

>经测试，第二版性能不如第一版，内存开销也比第一版大，原因是Task Channel的维护需要占用内存并且也有锁开销，而且众多goroutine从一个channel拿任务带来的竞争反而更加激烈。

```
package main

import (
	"gpool"

	"fmt"
	"runtime"
	"time"
)

func slow(n int) int {
	f0, f1 := 0, 1
	if n == 0 {
		return f0
	}
	if n == 1 {
		return f1
	}
	res := slow(n-2) + slow(n-1)
	return res
}

func quick() {
	n := 1000
	f0, f1 := 0, 1
	for i := 2; i <= n; i++ {
		f0, f1 = f1, f0+f1
	}
	fmt.Println(f1)
}

func withPool() {
	pool := gpool.NewPool(5)
	for i := 0; i < 20; i++ {
		pool.SubmitTask(func() { slow(32) })
	}

	for i := 0; i < 2; i++ {
		time.Sleep(time.Second)
		runtime.GC()
	}
	pool.Release()
	runtime.GC()
}

func noPool() {
	for i := 0; i < 20; i++ {
		go func() { slow(32) }()
	}

	for i := 0; i < 2; i++ {
		time.Sleep(time.Second)
		runtime.GC()
	}
	runtime.GC()
}

func main() {
	//noPool()
	withPool()
}

```

**使用协程池：**
```
SCHED 506ms: gomaxprocs=4 idleprocs=4 threads=6 spinningthreads=0 idlethreads=4 runqueue=0 gcwaiting=0 nmidlelocked=0 stopwait=0 sysmonwait=0
  P0: status=0 schedtick=10 syscalltick=1 m=-1 runqsize=0 gfreecnt=0
  P1: status=0 schedtick=11 syscalltick=2 m=-1 runqsize=0 gfreecnt=0
  P2: status=0 schedtick=11 syscalltick=1 m=-1 runqsize=0 gfreecnt=0
  P3: status=0 schedtick=9 syscalltick=0 m=-1 runqsize=0 gfreecnt=0
  M5: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M4: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M3: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M2: p=-1 curg=17 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M1: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 helpgc=0 spinning=false blocked=false lockedg=-1
  M0: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  G1: status=4(sleep) m=-1 lockedm=-1
  G2: status=4(force gc (idle)) m=-1 lockedm=-1
  G3: status=4(GC sweep wait) m=-1 lockedm=-1
  G4: status=4(finalizer wait) m=-1 lockedm=-1
  G5: status=4(chan receive) m=-1 lockedm=-1
  G6: status=4(chan receive) m=-1 lockedm=-1
  G7: status=4(chan receive) m=-1 lockedm=-1
  G8: status=4(chan receive) m=-1 lockedm=-1
  G9: status=4(chan receive) m=-1 lockedm=-1
  G10: status=4(chan receive) m=-1 lockedm=-1
  G17: status=3() m=2 lockedm=-1
```

**不使用协程池：**
```
SCHED 504ms: gomaxprocs=4 idleprocs=4 threads=6 spinningthreads=0 idlethreads=4 runqueue=0 gcwaiting=0 nmidlelocked=0 stopwait=0 sysmonwait=0
  P0: status=0 schedtick=9 syscalltick=2 m=-1 runqsize=0 gfreecnt=1
  P1: status=0 schedtick=13 syscalltick=0 m=-1 runqsize=0 gfreecnt=6
  P2: status=0 schedtick=12 syscalltick=0 m=-1 runqsize=0 gfreecnt=6
  P3: status=0 schedtick=11 syscalltick=0 m=-1 runqsize=0 gfreecnt=7
  M5: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M4: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M3: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M2: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  M1: p=-1 curg=-1 mallocing=0 throwing=0 preemptoff= locks=1 dying=0 helpgc=0 spinning=false blocked=false lockedg=-1
  M0: p=-1 curg=25 mallocing=0 throwing=0 preemptoff= locks=0 dying=0 helpgc=0 spinning=false blocked=true lockedg=-1
  G1: status=4(sleep) m=-1 lockedm=-1
  G2: status=4(force gc (idle)) m=-1 lockedm=-1
  G3: status=4(GC sweep wait) m=-1 lockedm=-1
  G4: status=4(finalizer wait) m=-1 lockedm=-1
  G5: status=6() m=-1 lockedm=-1
  G6: status=6() m=-1 lockedm=-1
  G7: status=6() m=-1 lockedm=-1
  G8: status=6() m=-1 lockedm=-1
  G9: status=6() m=-1 lockedm=-1
  G10: status=6() m=-1 lockedm=-1
  G11: status=6() m=-1 lockedm=-1
  G12: status=6() m=-1 lockedm=-1
  G13: status=6() m=-1 lockedm=-1
  G14: status=6() m=-1 lockedm=-1
  G15: status=6() m=-1 lockedm=-1
  G16: status=6() m=-1 lockedm=-1
  G17: status=6() m=-1 lockedm=-1
  G18: status=6() m=-1 lockedm=-1
  G19: status=6() m=-1 lockedm=-1
  G20: status=6() m=-1 lockedm=-1
  G21: status=6() m=-1 lockedm=-1
  G22: status=6() m=-1 lockedm=-1
  G23: status=6() m=-1 lockedm=-1
  G24: status=6() m=-1 lockedm=-1
  G25: status=3() m=0 lockedm=-1

```
**已经成功复用协程，第一步目标完成~**
## 第二步：优化性能 ##

 

1. 使用channel信号通知空闲worker到来，开销太大，参考ants，改为sync.Cond接受worker被放回空闲队列的信号；
2. worker使用无缓冲任务channel，为避免channel阻塞改为使用有缓冲的任务channel，缓冲容量设为1;

>考虑中：目前执行一个任务就被放回队列，而放回队列需要加锁。可尝试控制任务队列剩余任务数量达到阈值才将worker放回队列。

## 第三步：探索HttpServer工作流程，加入协程池 ##
>go version go1.9.2 linux/amd64

>net/http/server.go
>2970:http.ListenAndServe
>2705:Server.ListenAndServe
>2756:Server.Serve
>2798:go *http.conn.serve 这里开一个goroutine处理accept的连接
>1719:*conn.serve
>2689:serverHandler.SereHTTP
>2331:DefaultServeMux.ServeHTTP
>h = mux.Handler(r)2275:根据path匹配handler
>h.ServeHttp(w,r):匹配到经由http.Handle/http.HandleFunc注册的handler执行对应的函数


**探索过程中遇到一个问题：**
>如果想要控制HttpServer并发的goroutine数量应该在2798行go c.serve()之前控制c.serve()的执行方式，
ants给出的示例中，通过HandleFunc中的匿名函数使用协程池复用协程,相当于走到最后一步再使用协程池是不是有问题？

后来想明白了：减少协程数量不是使用协程池的根本出发点，复用协程池的内存资源，减少系统整体内存资源消耗才是使用协程池的根本原因。
所以，哪怕每个请求过来还是会开一个goroutine来处理，但是很快就进入到协程池中的某个goroutine运行，避免了额外的栈增长，从而提高了性能，优化了内存使用率。
