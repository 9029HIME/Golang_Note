package main

type ListNode struct {
	Val  int
	Next *ListNode
}

//解法1 通过递归
func swapPairs(head *ListNode) *ListNode {
	// 作为一个完整的链表，返回给上层使用
	if head == nil || head.Next == nil {
		return head
	}
	/**
	本级递归需要做的事
	1.将next.next递归下去，获取递归后的完整链表
	2.head和head.Next调换位置，同时将调换位置后的head指向"递归后的完整链表"的头节点
	3.将调换位置后的head.Next作为完整链表返回给上一级递归
	*/
	next := head.Next
	head.Next = swapPairs(head.Next.Next)
	next.Next = head
	return next
}
