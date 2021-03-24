package main

import "fmt"

/**
题目：
	编写一个函数来查找字符串数组中的最长公共前缀。如果不存在公共前缀，返回空字符串 ""。
结果：
	O(m×n)
*/
func main() {
	strs := []string{"ab", "a"}
	fmt.Println(prefix(strs))
	//fmt.Println("dog"[:1])
}

func prefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}

	if len(strs) == 1 {
		return strs[0]
	}

	//先假设第一个就是最长前缀了
	result := strs[0]

	for i := 1; i < len(strs); i++ {
		next := strs[i]
		length := min(len(result), len(next))
		var index int
		for j := 0; j < length; j++ {
			if result[j] != next[j] {
				break
			}
			index++
		}
		//到了这里，就能发现本字符串与下一个字符串的最长匹配长度，为index+1
		result = next[:index]
	}

	return result
}

func min(a int, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
