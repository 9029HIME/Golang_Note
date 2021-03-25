package main

import "math"

func main() {
	println(reverse(120))
}

/**
给你一个 32 位的有符号整数 x ，返回将 x 中的数字部分反转后的结果。

如果反转后整数超过 32 位的有符号整数的范围 [−2的31次方,  2的31 − 1] ，就返回 0。

假设环境不允许存储 64 位整数（有符号或无符号）。

结果：O(n)
核心：通过x%10获取尾数，x/10获取去尾数的数字，再通过 [(结果×10) + 尾数]的方式拼接结果
*/

func reverse(x int) int {
	if x == 0 {
		return 0
	}

	//看看要不要变为负数
	var isPositive bool = true
	if x < 0 {
		isPositive = false
		x = 0 - x
	}

	var result int

	for x > 0 {
		// 求x剩余的最后一位数字 如123 % 10 = 3
		last := x % 10
		// x/10可以获得去掉最后一位数的数，如123/10 = 12，当第一位数1/10结果为0处理完时就不再进入循环
		x = x / 10
		// 上面这两步相当于把3从123里分离，获得12和3

		// 接下来是将最后一位数字放到合适的位置，result的初始值为0，所以第一次循环的结果是3，下一次的循环是(3 * 10) + 尾数 = 32，以此类推
		result = (result * 10) + last
	}

	if result < math.MinInt32 || result > math.MaxInt32 {
		return 0
	}

	if !isPositive {
		result = 0 - result
	}

	return result

}
