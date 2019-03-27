package main

import (
	"pool/gpool"

	"fmt"
	"runtime"
	"time"
)

type task struct {
	f   func(int) int
	arg int
}

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

func quick(n int) int {
	f0, f1 := 0, 1
	for i := 2; i <= n; i++ {
		f0, f1 = f1, f0+f1
	}
	fmt.Println(f1)
	return f1
}

func (t task) Run() {
	fmt.Println(t.f(t.arg))
}

func main() {
	pool := gpool.NewPool(5, 10)
	for i := 0; i < 15; i++ {
		pool.SubmitTask(task{f: quick, arg: 1000})
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Second)
		runtime.GC()
	}

}
