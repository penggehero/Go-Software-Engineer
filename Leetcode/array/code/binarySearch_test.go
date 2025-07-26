package code

// binarySearch1 二分查找
// 第一种写法，我们定义 target 是在一个在左闭右闭的区间里，也就是[left, right]
// [left,right]
func binarySearch1(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for right >= left {
		middle := left + (right-left)/2
		if nums[middle] == target {
			return middle
		} else if nums[middle] > target {
			right = middle - 1
		} else if nums[middle] < target {
			left = middle + 1
		}
	}
	return -1000
}

// binarySearch2 二分查找
// 第二种写法，我们定义 target 是在一个在左闭右闭的区间里，也就是[left, right)
// [left,right]
func binarySearch2(nums []int, target int) int {
	left, right := 0, len(nums)
	for right > left {
		middle := left + (right-left)/2
		if nums[middle] == target {
			return middle
		} else if nums[middle] > target {
			right = middle
		} else if nums[middle] < target {
			left = middle + 1
		}
	}
	return -1
}
