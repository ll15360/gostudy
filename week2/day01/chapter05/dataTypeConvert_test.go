package chapter05

import (
	"strconv"
	"testing"
)

func TestConvert(t *testing.T) {
	//基础的数值类型转换，float,int,byte,rune等是可以直接转换的，前提是要保证转换后类型的范围能够容纳转换前的值，否则就会发生溢出，导致精度丢失。
	// 关于类型转换，在go中，没有隐式转换，强制类型转换需要注意精度丢失问题
	f1 := 3.1415926
	r := 2.356
	t.Log(f1 * r * r)

	// 例如这个 int16 转 int8 ，超过范围就会精度丢失
	var a int = 255
	b := int8(a)
	t.Log(b)

	// 浮点数转换为证书，精度也会丢失
	f2 := 3.14
	f3 := int(f2)
	t.Log(f3)

	// 关于byte 和rune的差别，本质上 byte是uint8,rune为uint32，rune表示的是一整个字符，而byte表示的一个字节(对于ascall码值英文字母和数字都是用byte表示，但是中文需要用三个字节表示会被截断乱码)
	s := "hello,你好"
	bytes := []byte(s)
	t.Log(len(bytes)) //会输出5+1+6=12个字节
	runes := []rune(s)
	t.Log(len(runes)) //会输出8个字符

	// bool值的转换，布尔值不能直接转换，可以通过对比
	// b1 := bool(s) 这是不允许的
	// i := int(true) 这同样是不允许的
	b1 := f2 > 4
	t.Log(b1)

	// 字符串与基本类型如何转换呢 借助strconv
	str := "123"
	c, err := strconv.Atoi(str)
	if err != nil {
		t.Log("类型不匹配，转换错误")
	} else {
		t.Log(c)
	}

	num := 666
	str1 := strconv.Itoa(num)
	t.Log(str1)

	// 其余的基本类型也是
	sf := strconv.FormatFloat(3.1415926, 'f', 2, 32)
	t.Log(sf)
	str2 := "3.1415926"
	sf1, err := strconv.ParseFloat(str2, 64)
	if err != nil {
		t.Log("类型转换错误,非浮点型字符串")
	} else {
		t.Log(sf1)
	}

}
