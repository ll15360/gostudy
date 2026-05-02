package chapter06

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

// 在go中处理异常，没有像java中的try catch监听一整段代码一样，但是可以通过 一行简写和函数封装达到

func convertStringtoInt(str string) (int, error) {
	// 字符串转数值类型可能出现异常
	i, err := strconv.Atoi(str)
	if err != nil {
		return -1, err
	}
	return i, nil
}

// 自定义函数：返回 结果+错误
func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("除数不能为0")
	}
	return a / b, nil
}

func TestError(t *testing.T) {
	// 在go的体系中,异常处理通常有两套核心机制,error 和 panic（通常与recover）
	i, err := convertStringtoInt("123")
	if err != nil {
		t.Log("类型转换异常")
	} else {
		t.Log(i)
	}

	ans, err := divide(10, 0)
	if err != nil {
		t.Log("除数为零")
	}
	t.Log(ans)

}

func ArrayOutBoundary() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("捕获到异常", err)
		}
	}()

	/// 等价 try 块（崩溃代码）
	var arr []int
	print(arr[100]) // 数组越界 panic
}

func TestPanic(t *testing.T) {
	// 这里先介绍一个defer函数 :延迟函数，当其所在的函数执行完或者崩溃时会执行这个defer函数，主要用于关闭资源,相当于finally
	ArrayOutBoundary()
}
