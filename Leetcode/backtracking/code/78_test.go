package code

/*
78. 子集

给你一个整数数组 nums ，数组中的元素 互不相同 。返回该数组所有可能的子集（幂集）。
解集 不能 包含重复的子集。你可以按 任意顺序 返回解集。

示例 1：
输入：nums = [1,2,3]
输出：[[],[1],[2],[1,2],[3],[1,3],[2,3],[1,2,3]]

示例 2：
输入：nums = [0]
输出：[[],[0]]

*/

// 子集问题是求树的全部节点
func subsets(nums []int) [][]int {
	var ret [][]int
	var arr []int
	n := len(nums)
	var dfs func(index int)
	dfs = func(index int) {
		// 每次选择，把路径保存
		ret = append(ret, append([]int{}, arr...))
		if index >= n {
			return
		}
		for i := index; i < n; i++ {
			// 选i
			arr = append(arr, nums[i])

			// 选下一个位置
			dfs(i + 1)

			// 撤销选i
			arr = arr[:len(arr)-1]
		}
	}
	dfs(0)
	return ret

}

func subsets2(nums []int) [][]int {
	var ans [][]int
	var t []int

	var dfs func(cur int)
	dfs = func(cur int) {
		if cur == len(nums) {
			ans = append(ans, append([]int{}, t...))
			return
		}

		// 选择当前元素
		t = append(t, nums[cur])
		dfs(cur + 1)

		// 不选择当前元素
		t = t[:len(t)-1]
		dfs(cur + 1)
	}

	dfs(0)
	return ans
}
