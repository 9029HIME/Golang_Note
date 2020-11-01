package main

func main() {

}

type ListNode struct {
	Val  int
	Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	var result *ListNode
	var tail *ListNode
	surplus := 0

	for l1 != nil || l2 != nil {
		var num1, num2 int
		if l1 == nil {
			num1 = 0
		} else {
			num1 = l1.Val
			l1 = l1.Next
		}
		if l2 == nil {
			num2 = 0
		} else {
			num2 = l2.Val
			l2 = l2.Next
		}
		plus := num1 + num2
		if surplus > 0 {
			plus = plus + surplus
			surplus = 0
		}
		surplus = plus / 10
		if surplus > 0 {
			plus = plus % 10
		}
		// newNode := new(ListNode)
		// newNode.Val = plus
		// newNode.Next = stack
		// stack = newNode
		if result == nil {
			result = new(ListNode)
			result.Val = plus
			tail = result
		} else {
			tail.Next = &ListNode{Val: plus}
			tail = tail.Next
		}
	}
	if surplus > 0 {
		tail.Next = &ListNode{Val: surplus}
		tail = tail.Next
	}
	return result
}
