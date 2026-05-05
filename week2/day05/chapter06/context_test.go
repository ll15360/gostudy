package chapter06

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// context的四个功能

// context.Context 是 Go 标准库用来控制协程生命周期的上下文对象。
// 2. 四大核心能力
// 取消信号：主动通知子协程停止工作
// 超时控制：自动到时间终止任务
// 截止时间：指定某个时间点必须结束
// 跨协程传元数据：链路追踪 ID、租户、请求标识

// 设计思想:父协程通过这个context控制所有子协程
func TestContextBase(t *testing.T) {
	// 根上下文 context.Background() 和 占位符上下文 context.TODO()
	ctx1 := context.Background()
	ctx2 := context.TODO()

	t.Logf("Background: %v", ctx1)
	t.Logf("TODO: %v", ctx2)

	// 根上下文默认没有取消，没有截止时间
	t.Log("是否已取消：", ctx1.Done())
	// ctx1.Deadline() 返回两个值 (time.Time, bool)，需要先解包再打印
	d, ok := ctx1.Deadline()
	t.Log("截止时间：", d, ok)
}

// 四大派生上下文：context.WithCancel 手动取消

func workerCancel(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			{
				fmt.Println("收到取消，退出协程", name)
				return
			}
		default:
			{
				fmt.Println("正在运行中", name)
			}
		}
	}
}
func TestContextWithCancel(t *testing.T) {
	// 定义根上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go workerCancel(ctx, "工作线程1")
	go workerCancel(ctx, "工作线程2")
	time.Sleep(2 * time.Second)
	cancel()
	time.Sleep(500 * time.Millisecond)
	// 小答疑:对于select监听时，就绪事件会触发case代码片段的执行，一类是管道有序数据就绪了，一类是管道关闭了，也会触发
	// 本质上 case ctx.Done()返回一个只读的空管道，调用这个cancel()会关闭这个空管道，从而执行那个case逻辑:还是管道通信
}

// 超时关闭
func contextWithTimeOut(ctx context.Context, name string) {
	for {
		select {
		case <-ctx.Done():
			{
				fmt.Println("退出线程", name)
				return
			}
		default:
			{
				fmt.Println("正常执行", name)
				time.Sleep(time.Millisecond * 500)
			}
		}
	}
}
func TestContextWithTimeOut(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	go contextWithTimeOut(ctx, "工作线程1")
	go contextWithTimeOut(ctx, "工作线程2")
	go contextWithTimeOut(ctx, "工作线程3")
	time.Sleep(3 * time.Second)

}

// 通过这个context传值
// 自定义Key（防止冲突）
type key string

const TraceIDKey key = "trace_id"

// 子函数：从ctx取值
func subFunc(ctx context.Context) {
	// 从上下文拿值
	traceID, ok := ctx.Value(TraceIDKey).(string)
	if !ok {
		fmt.Println("未获取到TraceID")
		return
	}
	fmt.Printf("子函数获取TraceID：%s\n", traceID)
}

func TestContextWithValue(t *testing.T) {
	// 1. 往ctx存值（根ctx + 键 + 值）
	ctx := context.WithValue(context.Background(), TraceIDKey, "K8S-TRACE-10086")

	// 2. 传递给子函数/子协程
	subFunc(ctx)

	// 也可以传给协程
	go func(ctx context.Context) {
		traceID := ctx.Value(TraceIDKey)
		fmt.Printf("协程获取TraceID：%s\n", traceID)
	}(ctx)

	time.Sleep(1 * time.Second)
}
