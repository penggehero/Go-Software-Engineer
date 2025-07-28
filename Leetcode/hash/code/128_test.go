package code

import (
	"slices"
)

/*
128. 最长连续序列
给定一个未排序的整数数组 nums ，找出数字连续的最长序列（不要求序列元素在原数组中连续）的长度。
请你设计并实现时间复杂度为 O(n) 的算法解决此问题。

示例 1：
输入：nums = [100,4,200,1,3,2]
输出：4
解释：最长数字连续序列是 [1, 2, 3, 4]。它的长度为 4。

示例 2：
输入：nums = [0,3,7,2,5,8,4,6,0,1]
输出：9

示例 3：
输入：nums = [1,0,1,2]
输出：3
*/

// longestConsecutive 排序
func longestConsecutive(nums []int) int {
	n := len(nums)
	if n < 2 {
		return n
	}
	slices.Sort(nums)
	ret := 1
	tempLength := ret
	for i := 1; i < n; i++ {
		// 元素相同，往后滑动
		if nums[i] == nums[i-1] {
			continue
		}
		// 连续则++
		if nums[i] == nums[i-1]+1 {
			tempLength++
			ret = max(tempLength, ret)
		} else {
			// 不连续，重置为1
			tempLength = 1
		}
	}
	return ret

}

// longestConsecutive 哈希
func longestConsecutive2(nums []int) int {
	numSet := map[int]bool{}
	for _, num := range nums {
		numSet[num] = true
	}
	ret := 0
	for num := range numSet {
		// 如果这个数是连续数列的起点
		// 则这个点是没有先驱元素的
		if !numSet[num-1] {
			currentNum := num
			length := 1
			// 遍历该起点的所有元素
			for numSet[currentNum+1] {
				// 元素改成下一个元素
				currentNum++
				// 长度+1
				length++
			}
			ret = max(ret, length)
		}
	}
	return ret
}
