package chapter02

// 在go语言中，有以下几种数据类型：
// 整数类型：int, int8, int16, int32, int64
// 浮点数类型：float32, float64
// 字符类型：byte
// 字符串类型：string
// 布尔类型：bool
// 空类型：nil

type Person struct {
	FirstName string
	LastName  string
	Age       int
}

type Shape interface {
	Area() float64
}

type Circle struct {
	Radius float64
}

func (c *Circle) Area() float64 {
	return Pi * c.Radius * c.Radius
}

func (c *Circle) Perimeter() float64 {
	return 2 * Pi * c.Radius
}

func add(a, b int) int {
	return a + b
}
