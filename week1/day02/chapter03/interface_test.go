package chapter03

import (
	"fmt"
	"testing"
)

// 在这个文件中，会探索go的多态是如何实现,当一个结构体实现某个接口的所有方法就会自动属于该结构类型

// go的多态分为三步 结构体实现接口方法(类型提升为接口类型),定义一个方法接收接口类型，调用时传入不同的结构体实例，自动调用对应的实现方法

// 定义三个动物类
type Dog struct{}

type Cat struct{}

type Bird struct{}

// 接口的定义:有方法签名，但是无具体的实现方法
type Animal interface {
	Speaker() string
}

func (d Dog) Speaker() string {
	return "汪汪汪"
}

func (c Cat) Speaker() string {
	return "喵喵喵"
}

func (b Bird) Speaker() string {
	return "叽叽叽"
}

func makeSound(a Animal) {
	fmt.Println(a.Speaker())
}

func TestInterface(t *testing.T) {
	d := &Dog{}
	makeSound(d)
}
