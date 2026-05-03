package chapter06

import (
	"fmt"
	"sync"
	"testing"
)

var conut = 0

var lock sync.Mutex

var condA = sync.NewCond(&lock)

var condB = sync.NewCond(&lock)

var condC = sync.NewCond(&lock)

func add(wg *sync.WaitGroup) {
	lock.Lock()
	defer wg.Done()
	defer lock.Unlock()
	conut++
}

func printA(wg *sync.WaitGroup) {
	lock.Lock()
	defer lock.Unlock()
	defer wg.Done()
	for conut < 99 {
		if conut%3 == 0 {
			fmt.Println("打印了A")
			condB.Signal()
			conut++
		} else {
			condA.Wait()
		}
	}

}

func printB(wg *sync.WaitGroup) {
	lock.Lock()
	defer lock.Unlock()
	defer wg.Done()
	for conut < 99 {
		if conut%3 == 1 {
			fmt.Println("打印了B")
			condC.Signal()
			conut++
		} else {
			condB.Wait()
		}
	}

}

func printC(wg *sync.WaitGroup) {
	lock.Lock()
	defer lock.Unlock()
	defer wg.Done()
	for conut < 99 {
		if conut%3 == 2 {
			fmt.Println("打印了C")
			condA.Signal()
			conut++
		} else {
			condC.Wait()
		}
	}

}

func TestMutex(t *testing.T) {
	// 在 go 中并发编程可以通过 mutex锁和channal进行通信,这里简单介绍go 的携程和 mutext
	var wg sync.WaitGroup
	// wg.Add(1000)
	// for i := 0; i < 1000; i++ {
	// 	go add(&wg)
	// }
	// wg.Wait()
	// t.Log(conut)
	wg.Add(3)
	go printA(&wg)
	go printB(&wg)
	go printC(&wg)
	wg.Wait()
}
