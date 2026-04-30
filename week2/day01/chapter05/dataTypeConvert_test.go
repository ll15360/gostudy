package chapter05

import (
	"testing"
)

func TestConvert(t *testing.T) {
	// 关于类型转换，在go中，没有隐式转换，强制类型转换需要注意精度丢失问题
	f1 := 3.1415926
	r := 2.356
	t.Log(f1 * r * r)

	// 例如这个 int16 转 int8 ，超过范围就会精度丢失
	var a int = 255
	b := int8(a)
	t.Log(b)

}
