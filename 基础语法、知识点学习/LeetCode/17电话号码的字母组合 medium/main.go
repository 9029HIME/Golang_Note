package main

import "fmt"

/**
给定一个仅包含数字 2-9 的字符串，返回所有它能表示的字母组合。答案可以按 任意顺序 返回。

给出数字到字母的映射如下（与电话按键相同）。注意 1 不对应任何字母。

来源：力扣（LeetCode）
链接：https://leetcode-cn.com/problems/letter-combinations-of-a-phone-number
著作权归领扣网络所有。商业转载请联系官方授权，非商业转载请注明出处。
*/

func main() {
	result := letterCombinations("234")
	for i, v := range result {
		fmt.Println(i, " : ", v)
	}
}

var phoneMap map[string]string = map[string]string{
	"2": "abc",
	"3": "def",
	"4": "ghi",
	"5": "jkl",
	"6": "mno",
	"7": "pqrs",
	"8": "tuv",
	"9": "wxyz",
}

/**
回溯每一种可能，定义一个指针nextDigits，用来指向下一个待回溯的数字，遍历该数字的所有可能， 并对该数字的所有可能进行回溯， 直到nextDigits指向的数字为空
*/

func letterCombinations(digits string) []string {
	combination := []string{}

	if digits == "" || digits == "1" {
		return combination
	}

	start := ""

	nextDigit := 0

	doCombination(start, nextDigit, digits, &combination)

	return combination
}

func doCombination(start string, nextDigit int, digits string, combinations *[]string) {
	// 直到nextDigits指向的数字为空
	len := len(digits)
	if nextDigit >= len {
		*combinations = append(*combinations, start)
		return
	}

	// 获取下一个待回溯的数字，进行组合
	next := string(digits[nextDigit])

	// 获取数组的英文内容
	letters := phoneMap[next]
	if letters == "" {
		return
	}

	for _, v := range letters {
		combination := start + string(v)
		// combination就会作为下一个start，继续匹配下去
		if nextDigit < len {
			nextNextDigit := nextDigit + 1
			nextStart := combination
			doCombination(nextStart, nextNextDigit, digits, combinations)
		} else {
			// 已经匹配完了，直接添加
			*combinations = append(*combinations, combination)
		}
	}
}
