package main

import "fmt"

/**
数字 n 代表生成括号的对数，请你设计一个函数，用于能够生成所有可能的并且 有效的 括号组合。
输入：n = 3
输出：["((()))","(()())","(())()","()(())","()()()"]
核心思路：
	1.把括号分成两批：左括号与有括号，且两批括号数都为3。
	2.回溯一切可能，只有当左括号数≤有括号数时，条件才成立。如果左括号数为3，右括号数为2，则说明情况是：)，必然不合法
	3.将条件成立的可能扔到数组里
	4.本题的关键是：回溯 + 条件成立
*/
func main() {
	parenthesis := generateParenthesis(3)
	fmt.Println(parenthesis)
}

func generateParenthesis(n int) []string {
	result := make([]string, 0)
	// 左括号数
	left := n
	// 右括号数
	right := n
	solution := ""
	doMatch(left, right, solution, &result)
	return result
}

//要记得切片扩容后，底层数组的指针值会变
func doMatch(left int, right int, solution string, result *[]string) {
	// 来到这里就要先判断条件是否成立？
	if left < 0 || right < 0 {
		return
	}

	if left > right {
		return
	}
	// 如果左右都为零，则说明已成立，往切片里加数据
	if left == 0 && right == 0 {
		*result = append(*result, solution)
		return
	}

	// 只有两种情况：放左括号、放右括号
	// 放左括号
	if left > 0 {
		/**
		left = left - 1
		TODO 注意这里千万不要改写left和right的值然后传进去回溯，否则回不到上一个状态，会重复多一次回溯
			比如((()))后，会回到(((的状态
		*/

		solution = solution + "("
		doMatch(left-1, right, solution, result)
		solution = solution[:len(solution)-1]
	}

	if right > 0 {
		// 放右括号
		//right = right - 1
		solution = solution + ")"
		doMatch(left, right-1, solution, result)
		solution = solution[:len(solution)-1]

	}
}
