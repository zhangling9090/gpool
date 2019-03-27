package main

import (
	"math/rand"
	"pool/gpool"
	"sync"
	"testing"
)

type testtask struct {
	f  func(int) int
	wg *sync.WaitGroup
}

func slowfib(n int) int {
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

func quickfib(n int) int {
	f0, f1 := 0, 1
	for i := 2; i <= n; i++ {
		f0, f1 = f1, f0+f1
	}
	return f1
}

func (t *testtask) Run() {
	t.f(rand.Int() % 1000)
	t.wg.Done()
}

var n = 100000

func withpool() {
	var wg sync.WaitGroup
	pool := gpool.NewPool(10000, 5*10000)
	for i := 0; i < n; i++ {
		wg.Add(1)
		var fib *testtask = &testtask{f: quickfib, wg: &wg}
		pool.SubmitTask(fib)
	}

	wg.Wait()
}

func BenchmarkWithPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		withpool()
	}
}

func nopool() {
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		var fib *testtask = &testtask{f: quickfib, wg: &wg}
		go func() {
			fib.Run()
		}()
	}

	wg.Wait()
}

func BenchmarkNoPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		nopool()
	}
}
