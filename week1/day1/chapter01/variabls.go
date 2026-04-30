package chapter01

import "fmt"

// 短变量声明只能在函数内部使用，不能在包级别使用
// f := 30 // 这行代码会导致编译错误，因为短变量声明不能在包级别使用

func main() {
	// go 中声明一个变量的方式，var 和:=
	var a int = 10
	fmt.Println("a:", a)

	// 使用 := 声明并初始化变量，编译器会根据右侧的值自动推断类型
	b := 20

	fmt.Println("b:", b)

	// var 声明变量时，如果没有初始化，变量会被赋予默认值（零值）,零值声明，不同类型的零值不同，int 的零值是 0，string 的零值是 ""，bool 的零值是 false 等等(零值)
	var c int
	fmt.Println("c (default value):", c)

	var d string
	fmt.Println("d (default value):", d)

	var e bool
	fmt.Println("e (default value):", e)

	// 正确的做法是在函数内部使用短变量声明
	f := 30
	fmt.Println("f:", f)

	// 测试提交

}
