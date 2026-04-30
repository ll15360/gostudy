package chapter01

import (
	"gostudy/week1/day1/chapter02"
	"testing"
)

func TestFibList(t *testing.T) {
	first := 1
	second := 1
	t.Log(first)
	for i := 0; i < 5; i++ {
		t.Log(" ", second)
		next := first
		first = second
		second = next + first
	}
}

func TestSwap(t *testing.T) {
	first := 1
	second := 2
	t.Log(first, second)

	first, second = second, first
	t.Log(first, second)
}

func TestPi(t *testing.T) {
	t.Log(chapter02.Pi)
}
