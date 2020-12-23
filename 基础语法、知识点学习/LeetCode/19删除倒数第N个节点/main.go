package main

func main() {

}

type ListNode struct {
	Val  int
	Next *ListNode
}

func removeNthFromEnd(head *ListNode, n int) *ListNode {
	slowNode := head
	fastNode := head

	for i := 0; i < n; i++ {
		fastNode = fastNode.Next
		if fastNode == nil {
			return doRemoveHead(slowNode)
		}
	}

	for fastNode.Next != nil {
		slowNode = slowNode.Next
		fastNode = fastNode.Next
	}
	doRemoveNext(slowNode)
	return head
}

func doRemoveNext(slowNode *ListNode) {
	removedNext := slowNode.Next
	realNext := slowNode.Next.Next
	slowNode.Next = realNext
	removedNext.Next = nil
}

func doRemoveHead(slowNode *ListNode) *ListNode {
	head := slowNode.Next
	slowNode.Next = nil
	return head
}
