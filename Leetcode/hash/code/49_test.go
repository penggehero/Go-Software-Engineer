package code

import (
	"slices"
)

/*
49. 字母异位词分组
给你一个字符串数组，请你将 字母异位词 组合在一起。可以按任意顺序返回结果列表。

示例 1:
输入: strs = ["eat", "tea", "tan", "ate", "nat", "bat"]
输出: [["bat"],["nat","tan"],["ate","eat","tea"]]
解释：
在 strs 中没有字符串可以通过重新排列来形成 "bat"。
字符串 "nat" 和 "tan" 是字母异位词，因为它们可以重新排列以形成彼此。
字符串 "ate" ，"eat" 和 "tea" 是字母异位词，因为它们可以重新排列以形成彼此。

示例 2:
输入: strs = [""]
输出: [[""]]

示例 3:
输入: strs = ["a"]
输出: [["a"]]
*/

// groupAnagrams
func groupAnagrams(strs []string) [][]string {
	m := map[string][]string{}
	for _, str := range strs {
		key := helper(str)
		arr, ok := m[key]
		if !ok {
			m[key] = []string{str}
		} else {
			m[key] = append(arr, str)
		}
	}
	var ret [][]string
	for _, arr := range m {
		ret = append(ret, arr)
	}
	return ret
}

// helper 排序，异位词排序后一定str一样
func helper(str string) string {
	arr := []rune(str)
	slices.Sort(arr)
	return string(arr)
}
