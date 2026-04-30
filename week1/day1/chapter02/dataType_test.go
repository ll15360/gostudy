package chapter02

import (
	"testing"
)

func TestAdd(t *testing.T) {
	var a int = add(1, 2)
	t.Log(a)
}

func TestCircleArea(t *testing.T) {
	var c = &Circle{Radius: 5}
	t.Log(c.Area())

}

func TestCirclePerimeter(t *testing.T) {
	var c = &Circle{Radius: 5}
	t.Log(c.Perimeter())
}

func TestDataType(t *testing.T) {
	// 注意点 测试函数必须是 Test+Xxx(大写字母开头)

	// int 类型的定义
	var a int = 10
	t.Log(a)
	// 短声明符
	b := 10
	t.Log(b)
	// 零值声明
	var c int
	t.Log(c)

	//字符串的定义
	var name string = "张三"
	t.Log(name)

	adress := "武汉"
	t.Log(adress)

	// 零值声明,字符串默认是空串
	var email string
	t.Log(email)

	// 布尔值的声明
	var isEmpty bool = true
	t.Log(isEmpty)

	isThink := true
	t.Log(isThink)

	// 零值声明 默认为false
	var isRed bool
	t.Log(isRed)

	// go提供了单精度和双精度的浮点数
	var Pi float32 = 3.1415926
	t.Log(Pi)

	Radius := 5.25

	// 下面这一行会报错 因为对于短声明值来说 : 像1，100 默认是 int类型,像5.25默认是 float64
	// t.Log(Pi * Radius * Radius)

	// 需要强转 ，但是最好是统一精度类型 或者强转是最好是由 float32 -> float64,否则可能精度丢失，且go是没有隐式转换的
	t.Log(Radius * Radius * float64(Pi))
}

func TestArray(t *testing.T) {
	// 数组类型
	// 零值声明 默认全为0
	var arr1 [5]int
	t.Log(arr1)

	// 短声明
	arr2 := [5]int{1, 2, 3, 4, 5}
	t.Log(arr2)

	var arr3 = [...]int{1, 2, 3, 4}
	t.Log(len(arr3))
}

func TestSlice(t *testing.T) {
	//切片是 Go 为了解决数组「长度固定、拷贝开销大」设计的动态数组，是引用类型。
	// 切片声明:
	var slice = []int{1, 2, 3}
	t.Log(slice)

	// 使用make声明
	var slice2 = make([]int, 3, 5)
	t.Log(slice2)

	// 直接使用切片截取,截取范围为左闭右开
	var arr = [5]int{1, 2, 3, 4, 5}
	var slice3 = arr[1:4]
	t.Log(slice3)

	// 切片的增删改查
	// 添加元素只能使用这个append
	slice2 = append(slice2, 4)
	t.Log(slice2)

	slice3 = append(slice3, 5)
	t.Log(slice3)

	// 查询长度与容量
	t.Log(len(slice2))
	t.Log("容量为 ", cap(slice2))

	// 切片的浅拷贝与深拷贝 本质区别就是浅拷贝的引用指向同一个内存地址，修改会影响原有的切片
	slice4 := slice3
	slice4[1] = 100
	t.Log(slice3)

	// 深拷贝方式
	slice5 := make([]int, len(slice3))
	copy(slice5, slice3)
	slice5[1] = 99
	// 需要注意的是，拷贝时是不会进行扩容的，选择容量更小的
	t.Log(slice5)

}

func TestMap(t *testing.T) {
	// 使用make声明一个map，初始为空
	user := make(map[string]string)
	user["address"] = "武汉"
	user["name"] = "张三"
	t.Log(user["address"])

	// 字面量直接赋值
	student := map[string]int{}
	student["age"] = 18
	t.Log(student["age"])

	// s删除与修改
	delete(student, "age")
	t.Log(student["age"])
}
