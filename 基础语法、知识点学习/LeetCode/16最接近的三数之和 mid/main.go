package main

import (
	"math"
	"sort"
)

/**
给定一个包括 n 个整数的数组 nums 和 一个目标值 target。找出 nums 中的三个整数，使得它们的和与 target 最接近。返回这三个数的和。假定每组输入只存在唯一答案。
来源：力扣（LeetCode）
链接：https://leetcode-cn.com/problems/3sum-closest
著作权归领扣网络所有。商业转载请联系官方授权，非商业转载请注明出处。

结果:O(n²+n)
核心思路：
	先将数组从低到高排序，然后三指针。第一指针为第一个数据，第二指针为第一指针后一个数据，第三个指针指向最后一个数据。
	三个指针数值相加，与现有结果比较取最小值！！！！！
	如果相加结果比目标值大，则代表最大值（第三指针）大了，此时将第三指针往左移动。如果比目标值小，则代表第二小值（第二指针）小了，此时将第二指针往右移动
	直到第二指针与第三指针重叠，说明第一指针范围内的所有第二与第三指针的结果已计算过
	此时将第一指针往后移，继续与第二指针、第三指针相加....

	在上面的循环过程中，现有结果永远是最小的，因为只有相加结果＜现有结果时，现有结果才会覆盖

	注意，如果现有结果和目标值刚好匹配，则直接返回即可

*/

func main() {
	nums := []int{-1, 2, 1, -4}
	println(threeSumClosest(nums, 1))
}

func threeSumClosest(nums []int, target int) int {

	if nums == nil {
		return math.MaxInt64
	}

	if len(nums) < 3 {
		return math.MaxInt64
	}

	sort.Ints(nums)

	result := nums[0] + nums[1] + nums[len(nums)-1]

	for i, v := range nums {
		next := i + 1
		tail := len(nums) - 1
		// 保证第二指针与第三指针重叠后退出当前循环
		for next < tail {
			sum := v + nums[next] + nums[tail]
			if sum == target {
				return sum
			}
			// 和最低结果比较取最小值
			if abs(target-sum) < abs(target-result) {
				result = sum
			}
			// 算出来的结果比target大，就是说最大值大了，减少一点
			if sum > target {
				tail--
			}
			// 算出来的结果比target大，就是说最小值小了，增加一点
			if sum < target {
				next++
			}
		}
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	} else {
		return x
	}
}
