package code

/*
给定两个整数 n 和 k，返回范围 [1, n] 中所有可能的 k 个数的组合。
你可以按 任何顺序 返回答案。

示例 1：
输入：n = 4, k = 2
输出：
[
  [2,4],
  [3,4],
  [2,3],
  [1,2],
  [1,3],
  [1,4],
]

示例 2：
输入：n = 1, k = 1
输出：[[1]]
*/

// combine 不剪枝
func combine(n int, k int) [][]int {
	var res [][]int
	var dfs func(index int, length int, num int, arr []int)
	dfs = func(index int, length int, num int, arr []int) {
		if len(arr) == num {
			res = append(res, append([]int{}, arr...))
			return
		}
		for i := index; i <= length; i++ {
			// 选择
			arr = append(arr, i)
			// 递归
			dfs(i+1, length, num, arr)
			// 撤销
			arr = arr[:len(arr)-1]
		}
	}
	dfs(1, n, k, []int{})
	return res
}

var (
	path []int
	res  [][]int
)

func combine2(n int, k int) [][]int {
	path, res = make([]int, 0, k), make([][]int, 0)
	dfs(n, k, 1)
	return res
}

// dfs 剪枝优化
func dfs(n int, k int, start int) {
	if len(path) == k { // 说明已经满足了k个数的要求
		tmp := make([]int, k)
		copy(tmp, path)
		res = append(res, tmp)
		return
	}
	for i := start; i <= n; i++ { // 从start开始，不往回走，避免出现重复组合
		if n-i+1 < k-len(path) { // 剪枝
			break
		}
		path = append(path, i)
		dfs(n, k, i+1)
		path = path[:len(path)-1]
	}
}
