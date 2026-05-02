package chapter06

import (
	"sync"
	"testing"
)

var conut = 0

var lock sync.Mutex

func add(wg *sync.WaitGroup) {
	lock.Lock()
	defer wg.Done()
	defer lock.Unlock()
	conut++
}

func TestMutex(t *testing.T) {
	// 在 go 中并发编程可以通过 mutex锁和channal进行通信,这里简单介绍go 的携程和 mutext
	var wg sync.WaitGroup
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go add(&wg)
	}
	wg.Wait()
	t.Log(conut)
}
