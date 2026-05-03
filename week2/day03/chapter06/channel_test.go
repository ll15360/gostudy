package chapter06

// 核心思想 :不要使用共享内存来通信，要使用通信来共享内存

import (
	"fmt"
	"sync"
	"testing"
)

func send(ch chan string, message string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(ch)
	ch <- message
	fmt.Println("发送了一条消息: ", message)
	fmt.Println("发送方关闭管道")
}

func recive(ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var message, ok = <-ch
	if ok {
		fmt.Println("接受了一条消息: ", message)
	} else {
		fmt.Println("管道已经关闭")
	}
}

// 生产者与消费者模型
func producer(ch chan<- int) {
	defer close(ch) //注意管道的关闭
	for i := 0; i < 10; i++ {
		ch <- i
		fmt.Println("已经传递任务", i)
	}
}

func consunmer(id int, wg *sync.WaitGroup, ch <-chan int) {
	defer wg.Done()
	for task := range ch {
		fmt.Printf("消费者%d 处理任务：%d\n", id, task)
	}
}

func TestChannel(t *testing.T) {
	// 同步管道与异步管道 :本质就是 容量, 满则写阻塞，空则取阻塞
	// 发送方阻塞 → 直到接收方取数据(同步管道)
	ch := make(chan int, 5)
	var wg sync.WaitGroup
	// wg.Add(2)
	// go send(ch, "你好啊", &wg)
	// go recive(ch, &wg)
	// go recive(ch, &wg)
	wg.Add(3)
	go producer(ch)
	for i := 0; i < 3; i++ {
		go consunmer(i, &wg, ch)
	}
	wg.Wait()
	//单向管道: 为了规范化编程,只能规定:只接受，只发送
	t.Log("全部任务处理完成")
}
