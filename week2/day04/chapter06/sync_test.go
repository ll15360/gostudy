package chapter06

import (
	"sync"
	"testing"
)

// 这里主要是学习这个sync包

// sync.mutex锁，

var count int = 0
var mu sync.Mutex

// 通过mutex去控制共享变量的并发安全，但是更推荐channel写法
func TestMutex(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			mu.Lock()
			defer mu.Unlock()
			defer wg.Done()
			count++
		}()
	}
	wg.Wait()
	t.Log(count)
}

// 通道的写法
func TestChannelA(t *testing.T) {
	ch := make(chan int, 5)
	var wg sync.WaitGroup
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			ch <- 1
			defer wg.Done()
		}()
	}
	// 不建议在主协程中去关闭通道,应该如此写:
	go func() {
		wg.Wait()
		close(ch)
	}()
	// close(ch)
	//  在主线程收集结果
	for num := range ch {
		count += num
	}
	t.Log(count)
}

// go当中的单例模式
// 在java中，构造一个单例模式需要自己定义，go中提供了一个sync.Once，通过.do()传入一个只会执行一次的函数
type Person struct {
	Name string
	Age  int
}

var config string
var per *Person

func getPersonandConfig() {
	config = "某些需要单例的配置"
	per = &Person{
		Name: "zhangsan",
		Age:  18,
	}
}

func TestOnce(t *testing.T) {
	once := sync.Once{}
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		// 模拟十个线程并发创建
		go func(id int) {
			defer wg.Done()
			once.Do(getPersonandConfig)
			t.Log("协程id为：", id, "创建了config", &config, "创建了", per)
		}(i)
	}
	wg.Wait()
}

// map与sync.Map(并发安全)

// var mp = make(map[string]string)

// // 下面这段代码会fail error
// func TestMap(t *testing.T) {
// 	// 模仿100个协程并发插入
// 	t.Log(len(mp))
// 	var wg sync.WaitGroup
// 	wg.Add(100)
// 	for i := 0; i < 100; i++ {
// 		go func() {
// 			mp["zhangsan"] = "test"
// 		}()
// 	}
// 	wg.Wait()
// 	t.Log(len(mp))
// }

var mp sync.Map

// 下面这段代码会fail error
func TestMap(t *testing.T) {
	// 模仿100个协程并发插入
	t.Log()
	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			mp.Store("zhangsan", 100)
		}()
	}
	wg.Wait()
	t.Log(mp.Load("zhangsan"))
}
