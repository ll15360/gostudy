package chapter06

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// 这里主要是写go并发编程的一些练习

// 练习 1：并发计算 + 管道收集结果
// 企业场景
// 批量计算数据（如商品价格换算、数据统计），用协程并发提升效率，通过管道统一收集结果。
// 要求
// 启动 5 个协程，分别计算 1~5 的平方
// 用无缓冲管道接收计算结果
// 主协程打印所有结果

func producerNum(ch chan int, num int) {
	ch <- num * num
	fmt.Println("结果已经传入管道", num*num)
}

func particeOne() {
	// 定义同步管道
	ch := make(chan int)
	for i := 0; i < 5; i++ {
		go producerNum(ch, i)
	}
	// 由主线程读取管道结果,注意，如果不是主线程读取结果，需要通过wg.WaitGroup
	for i := 1; i <= 5; i++ {
		fmt.Printf("结果：%d\n", <-ch)
	}
	defer close(ch)
}

func TestParticeOne(t *testing.T) {
	particeOne()
}

// 练习2
// 基础生产者消费者模型
// 企业场景
// 日志收集、消息队列简化版、数据生产 - 消费解耦（企业最常用模型）。
// 要求
// 生产者：1 个协程生成 1~10 的数字，写入管道
// 消费者：1 个协程从管道读取数据并打印
// 生产完成后关闭管道，消费者自动退出

func basePoducer(ch chan int) {
	defer close(ch)
	for i := 0; i < 10; i++ {
		ch <- i
		fmt.Println("生产者生产了一个任务", i)
	}
}

func baseConsumer(ch chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range ch {
		fmt.Println("消费者处理了任务", task)
	}
}

func particeTwo() {
	ch := make(chan int)
	var wg sync.WaitGroup
	go basePoducer(ch)
	wg.Add(1)
	go baseConsumer(ch, &wg)
	wg.Wait()

}

func TestParticeTwo(t *testing.T) {
	particeTwo()
	t.Log("所有任务完成")
}

// 练习 3：令牌桶限流 + 并发任务
// 企业场景
// 核心高频场景：限制并发协程数量（防止协程爆炸，如接口请求、数据库操作限流）。
// 要求
// 执行 20 个任务，最多同时运行 3 个协程

// 用有缓冲管道实现令牌桶限流

func particeThree() {
	var wg sync.WaitGroup
	// 1. 令牌桶：缓冲大小=3 → 最多3个并发协程（空结构体不占内存，Go最佳实践）
	tokenCh := make(chan struct{}, 3)

	// 2. 20个任务
	for i := 1; i <= 20; i++ {
		taskNum := i
		wg.Add(1)

		// 启动协程执行任务
		go func() {
			defer wg.Done() // 任务完成，计数器-1

			tokenCh <- struct{}{}
			defer func() { <-tokenCh }()

			// 模拟业务执行（比如接口请求、数据库操作）
			fmt.Printf("任务[%d] 开始执行，当前并发数：%d\n", taskNum, len(tokenCh))
			time.Sleep(100 * time.Millisecond) // 模拟耗时
			fmt.Printf("任务[%d] 执行完成\n", taskNum)
		}()
	}

	// 🔥 必须加：主协程等待所有子协程执行完毕
	wg.Wait()
	close(tokenCh)
	fmt.Println("🎉 20个任务全部执行完成！")
}

func TestParticeThree(t *testing.T) {
	particeThree()
}

// 练习 4：并发任务 + 错误收集
// 企业场景
// 批量任务执行（如文件处理、数据导入），单个任务失败不影响其他任务，统一收集错误信息。
// 要求
// 10 个并发任务，随机模拟任务成功 / 失败
// 用管道收集所有错误，主协程打印

// 练习 5：WaitGroup + Channel 优雅等待任务
// 企业场景
// 替代 time.Sleep，精准等待所有协程执行完毕，再关闭管道。
// 核心考点
// sync.WaitGroup + channel 配合、优雅关闭管道

// 练习 6：批量接口并发请求聚合
// 企业场景
// 同时请求多个微服务接口（如用户信息、订单信息、商品信息），并发请求提升响应速度，统一聚合结果。
// 要求
// 并发请求 3 个模拟接口
// 主协程等待所有结果，聚合后输出

// 练习 7：支持优雅退出的并发任务
// 企业场景
// 服务运行中接收退出信号（如 Ctrl+C），安全停止所有协程，不丢失任务。
// 核心考点
// select 多路复用、退出信号、优雅关闭

// 练习 8：电商订单批量并发处理系统
// 企业场景
// 批量处理 100 个订单，要求：
// 限流：最多 10 个协程同时处理
// 错误收集：记录处理失败的订单
// 结果统计：统计成功 / 失败数量
// 优雅退出：支持手动停止服务
// 核心考点
// 限流、错误处理、并发统计、优雅退出、生产者消费者全融合
