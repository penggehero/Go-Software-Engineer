package code

import (
	"sort"
)

/*
56. 合并区间
以数组 intervals 表示若干个区间的集合，其中单个区间为 intervals[i] = [starti, endi] 。
请你合并所有重叠的区间，并返回 一个不重叠的区间数组，该数组需恰好覆盖输入中的所有区间 。

示例 1：
输入：intervals = [[1,3],[2,6],[8,10],[15,18]]
输出：[[1,6],[8,10],[15,18]]
解释：区间 [1,3] 和 [2,6] 重叠, 将它们合并为 [1,6].

示例 2：
输入：intervals = [[1,4],[4,5]]
输出：[[1,5]]
解释：区间 [1,4] 和 [4,5] 可被视为重叠区间。
*/

// merge 合并区间
// 对于两个区间[x1, y1],[x2, y2]
// 总共可以分为两种情况，一种是重叠的，另一种是不重叠的
// 发生重叠时，必然成立 x2 <= y1,
// 因为是排序后的，所以x2 >= x1,这个时候只需要比较y2和y1的大小。
// 还有一种情况是[x2,y2]是[x1,y1]的真子集，这个时候不需要更新
// 不发生重叠，那么 x2 > y1， 不需要更新
func merge(intervals [][]int) [][]int {

	// 先排序,按照升序排序
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	// 两个辅助x1,y1
	x1, y1 := intervals[0][0], intervals[0][1]
	var ans [][]int
	// 从第2个元素开始遍历
	for i := 1; i < len(intervals); i++ {
		x2 := intervals[i][0]
		y2 := intervals[i][1]

		// 重叠时，必然成立 x2 <= y1,
		if x2 <= y1 {
			y1 = max(y2, y1)
		} else {
			// 不重叠，直接添加x1和y1
			ans = append(ans, []int{x1, y1})
			// 然后向后移动x1和y1
			x1 = x2
			y1 = y2
		}
	}
	//然后根据我们搜集到的x1和y1进行更新
	ans = append(ans, []int{x1, y1})
	return ans
}
