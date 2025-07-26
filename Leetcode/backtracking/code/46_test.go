package code

/*
给定一个不含重复数字的数组 nums ，返回其 所有可能的全排列 。你可以 按任意顺序 返回答案。

示例 1：
输入：nums = [1,2,3]
输出：[[1,2,3],[1,3,2],[2,1,3],[2,3,1],[3,1,2],[3,2,1]]

示例 2：
输入：nums = [0,1]
输出：[[0,1],[1,0]]

示例 3：
输入：nums = [1]
输出：[[1]]
*/

func permute(nums []int) [][]int {
	var res [][]int
	var arr []int
	length, m := len(nums), map[int]bool{}
	var dfs func()
	dfs = func() {
		if len(arr) == length {
			res = append(res, append([]int{}, arr...))
			return
		}
		// 从第一个元素开始选择
		for i := 0; i < length; i++ {
			// 已经使用过，不能再选择
			if m[nums[i]] {
				continue
			}
			// 选择arr 和 map
			arr = append(arr, nums[i])
			m[nums[i]] = true
			// 递归
			dfs()
			// 撤销arr 和 map
			arr = arr[:len(arr)-1]
			m[nums[i]] = false
		}
	}
	dfs()
	return res
}
