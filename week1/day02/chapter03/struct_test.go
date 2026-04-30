package chapter03

// 这里主要是学习:结构体,定义，初始化，接口，如何实现组合，多态

import (
	"fmt"
	"testing"
)

// 在go中如何实现多态与组合(没有继承)，我们需要注意的是字段的访问范围，与包的字段类似，如果字段以小写开头则只能在当前结构体访问，即使组合也不能访问
// 方式1 嵌套组合

type person struct {
	Height float32
	Weight float32
}

type student struct {
	person
	Name string
	Age  int
}

// 在go中，通过结构体加绑定方法达到类似java类的效果,这里需要注意的是，值传递与地址传递
// 下面的方法展示了一个值传递
func (stu student) sayHello() {
	fmt.Println("大家好，我是:" + stu.Name)
}

// 如果这个方法需要修改传入的结构体实例,需要通过指针接收student实例的地址
func (stu *student) updateName(name string) {
	stu.Name = name
}

func (stu student) printstu() {
	fmt.Println("这个学生是: ", stu.Name, "身高是:", stu.Height, "体重是:", stu.Weight, "KG")
}

func TestStructInit(t *testing.T) {
	// 初始化
	// 取址符初始化，拿到的是对象的地址，但是user.Age时会自动解析引用
	// 在go中没有类似java中的构造函数,
	// 推荐使用
	user := &student{
		Name: "张三",
		Age:  18,
	}
	t.Log(user.Age)
	t.Log(user.Name)
	user.Name = "李四"
	t.Log(user.Name)

	stu1 := student{
		Name: "小明",
		Age:  18,
	}
	stu1.Name = "小红"
	t.Log(stu1.Name)

	// 第三种初始化的方法
	stu2 := new(student)
	stu2.Name = "小刚"
	t.Log(stu2.Name)
}

func TestBindFunc(t *testing.T) {
	stu := &student{
		Name: "小红",
		Age:  19,
	}
	stu.sayHello()
	stu.updateName("小明")
	stu.sayHello()
}

func TestComposition(t *testing.T) {
	// 通过组合实现类似java中的继承
	stu := &student{
		person: person{
			Height: 1.75,
			Weight: 60.2,
		},
		Name: "小明",
		Age:  20,
	}
	stu.printstu()
}
