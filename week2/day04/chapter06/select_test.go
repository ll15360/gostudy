package chapter06

import (
	"sync"
	"testing"
)

// 这里主要学习select ：go的IO多路复用
// 同时监听多个 channel，哪个通道就绪就执行哪个 case
// 实现超时控制、阻塞等待、事件分发、优雅退出
// 关键特性
// 每个 case 必须是 channel 读写操作
// 多个 case 同时就绪：随机选一个执行
// 加 default：变为非阻塞，没通道就绪直接走 default
// 不加 default：阻塞等待，直到某个 channel 就绪

func TestSelect(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)
	var wg sync.WaitGroup
	wg.Add(2)

	// 发送协程
	go func() {
		defer wg.Done()
		ch1 <- "Pod状态变更"
	}()
	go func() {
		defer wg.Done()
		ch2 <- "Nginx节点状态变更"
	}()

	go func() {
		wg.Wait()
		close(ch1)
		close(ch2)
	}()

	count := 0
	for count < 2 {
		select {
		case msg, ok := <-ch1:
			if ok {
				t.Log("ch1:", msg)
				count++
			}
		case msg, ok := <-ch2:
			if ok {
				t.Log("ch2:", msg)
				count++
			}
		}
	}

	// 思考下面的代码块会出现什么问题
	// for i := 0; i < 2; i++ {
	// 	select {
	// 	case msg := <-ch1:
	// 		t.Logf("ch1 接收消息：%s", msg)
	// 	case msg := <-ch2:
	// 		t.Logf("ch2 接收消息：%s", msg)
	// 	}
	// }

	// 问题是:select当多个事件就绪时，会随机选中一个就绪事件，如果两次随机选中第二个就绪事件，可能出现事件1一直写入阻塞

	t.Log("执行完成，无阻塞、无泄漏")
}

// 无限监听加检测管道状态
func TestSelectandChannel(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer close(ch1)
		ch1 <- "Pod状态变更"
	}()
	go func() {
		defer wg.Done()
		defer close(ch2)
		ch2 <- "Nginx节点状态变更"
	}()

	// 无限监听，通道关闭后自动退出
	for {
		select {
		case msg, ok := <-ch1:
			if !ok {
				ch1 = nil // nil通道会被select忽略
				break
			}
			t.Log(msg)
		case msg, ok := <-ch2:
			if !ok {
				ch2 = nil
				break
			}
			t.Log(msg)
		}
		// 两个通道都关闭，退出循环
		if ch1 == nil && ch2 == nil {
			break
		}
	}

	wg.Wait()
}
