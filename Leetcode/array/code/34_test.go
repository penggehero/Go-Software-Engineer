package code

/*
34. 在排序数组中查找元素的第一个和最后一个位置
给你一个按照非递减顺序排列的整数数组 nums，和一个目标值 target。请你找出给定目标值在数组中的开始位置和结束位置。
如果数组中不存在目标值 target，返回 [-1, -1]。
你必须设计并实现时间复杂度为 O(log n) 的算法解决此问题。

示例 1：
输入：nums = [5,7,7,8,8,10], target = 8
输出：[3,4]

示例 2：
输入：nums = [5,7,7,8,8,10], target = 6
输出：[-1,-1]

示例 3：
输入：nums = [], target = 0
输出：[-1,-1]
*/
// searchRange
func searchRange(nums []int, target int) []int {
	left := getLeftBorder(nums, target)
	right := getRightBorder(nums, target)
	if left == -2 || right == -2 {
		return []int{-1, -1}
	}
	// 当存在只要1个元素的情况
	if (right - left) > 1 {
		return []int{left + 1, right - 1}
	}
	return []int{-1, -1}
}

func getRightBorder(nums []int, target int) int {
	var rightBorder = -2
	left, right := 0, len(nums)-1
	for right >= left {
		middle := left + (right-left)>>1
		// 	等于的情况下，依然往右边边查找边界
		if nums[middle] <= target {
			left = middle + 1
			rightBorder = left
		} else {
			right = middle - 1
		}
	}
	return rightBorder
}

func getLeftBorder(nums []int, target int) int {
	var leftBorder = -2
	left, right := 0, len(nums)-1
	for right >= left {
		middle := left + (right-left)>>1
		// 等于的情况下，依然往左边查找边界
		if nums[middle] >= target {
			right = middle - 1
			leftBorder = right
		} else {
			left = middle + 1
		}
	}
	return leftBorder
}
