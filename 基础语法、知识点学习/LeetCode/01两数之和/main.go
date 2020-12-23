package main

func main() {

}

func twoSum(nums []int, target int) []int {
	hashmap := make(map[int]int, 10)
	for i, v := range nums {
		/**
		  通过哈希表快速查询是否有合适的值，没有则放进去。
		  等轮到其他值时，可以直接被其他值拿出来比对
		  **/
		value, ok := hashmap[target-v]
		if ok {
			return []int{value, i}
		} else {
			hashmap[v] = i
		}
	}
	return nil
}
