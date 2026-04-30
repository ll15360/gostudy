package chapter04

// 这个文件算是阶段性的练习，基础类型，结构体，映射，切片，数组的应用，若干个算法题

import (
	"testing"
)

// 两数之和
func twoSum(nums []int, target int) []int {
	idx := map[int]int{} // 创建一个空哈希表
	for j, x := range nums {
		if i, ok := idx[target-x]; ok {
			return []int{i, j}
		}
		idx[x] = j //
	}
	return nil // 不会执行到这里
}

// 无重复最长子串
func lengthOfLongestSubstring(s string) (ans int) {
	cnt := [128]int{} // 也可以用 map，这里为了效率用的数组
	left := 0
	for right, c := range s {
		cnt[c]++
		for cnt[c] > 1 { // 窗口内有重复字母
			cnt[s[left]]--
			left++
		}
		ans = max(ans, right-left+1) //
	}
	return
}

// 链表,练习结构体

type ListNode struct {
	Next *ListNode
	Val  int
}

// 循环链表
func hasCycle(head *ListNode) bool {
	slow, fast := head, head // 乌龟和兔子同时从起点出发
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if fast == slow {
			return true
		}
	}
	return false // 访问到了链表末尾，无环
}

func TestTwoSumSolution(t *testing.T) {
	nums := [...]int{2, 7, 11, 15}
	target := 9
	t.Log(twoSum(nums[:], target))
}

func TestTengthOfLongestSubstring(t *testing.T) {
	s := "abcabcbb"
	t.Log(lengthOfLongestSubstring(s))
}

func TestHasCycle(t *testing.T) {
	arr := []int{1, 2, 3, 4}
	var head, tail *ListNode

	// 尾插法创建链表
	for _, v := range arr {
		node := &ListNode{Val: v}
		if head == nil {
			head = node
			tail = node
		} else {
			tail.Next = node
			tail = node
		}
	}

	// 最后构造环
	tail.Next = head

	// 测试结果断言
	if hasCycle(head) != true {
		t.Error("预期有环，结果无环")
	}
}
