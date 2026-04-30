package chapter04

import (
	"fmt"
	"testing"
)

// 在这个文件，将会探讨流程控制相关的语句

func TestIf(t *testing.T) {
	age := 18
	if age > 18 {
		fmt.Println("年龄大于18岁")
	} else { //细节 else必须和大括号在同一行
		fmt.Println("年龄小于18岁")
	}

	score := 90.5
	if score < 60 {
		fmt.Println("不及格")
	} else if score < 80 {
		fmt.Println("一般")
	} else {
		fmt.Println("良好")
	}

	// 第三种用法，先声明变量再使用该变量判断

	if num := 10; num > 5 {
		fmt.Println("数值大于5")
	}

	// 在go中使用map的场景中,通常比较常用
	user := make(map[string]string)
	user["name"] = "张三"
	if name, ok := user["name"]; ok {
		fmt.Println("存在该字段", name)
	}

}

// 循环
func TestFor(t *testing.T) {
	// 格式：for 初始值; 条件; 增量 { 代码 }
	for i := 0; i < 5; i++ {
		fmt.Println("循环次数：", i)
	}

	// 类似while
	// 无限循环的简化版，等价于 while
	i := 0
	for i < 5 {
		fmt.Println(i)
		i++
	}

	// 遍历数组，切片
	// 遍历集合（数组、切片、map、字符串）
	nums := []int{10, 20, 30}
	for index, value := range nums {
		fmt.Printf("索引：%d，值：%d\n", index, value)
	}

	// 测试break 和continue
	// break 示例：i=3 时直接结束循环
	for i := 0; i < 5; i++ {
		if i == 3 {
			break
		}
		fmt.Println(i) // 输出 0 1 2
	}

	// continue 示例：跳过 i=2，继续循环
	for i := 0; i < 5; i++ {
		if i == 2 {
			continue
		}
		fmt.Println(i) // 输出 0 1 3 4
	}
}
