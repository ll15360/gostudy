package chapter05

// 这里的类型转换主要是关注 类似java中的object类型 interface {}空接口

//关于空接口的类型转换
import (
	"testing"
)

func TestConvertInterface(t *testing.T) {
	var obj interface{}
	obj = "123"
	t.Log(obj)
	// 把obj转换为string
	str, ok := obj.(string)
	if ok {
		t.Log(str)
	}

	var obj1 interface{} = "hello"
	num, ok := obj1.(int)
	if ok {
		t.Log(num)
	} else {
		t.Error("类型转换失败")
	}
}
