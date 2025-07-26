package code

/*

电话号码的字母组合
给定一个仅包含数字 2-9 的字符串，返回所有它能表示的字母组合。答案可以按 任意顺序 返回。
给出数字到字母的映射如下（与电话按键相同）。注意 1 不对应任何字母。

示例 1：
输入：digits = "23"
输出：["ad","ae","af","bd","be","bf","cd","ce","cf"]

示例 2：
输入：digits = ""
输出：[]

示例 3：
输入：digits = "2"
输出：["a","b","c"]
*/

func letterCombinations(digits string) []string {
	if len(digits) == 0 {
		return []string{}
	}

	letterMap := map[uint8][]string{
		'2': {"a", "b", "c"},
		'3': {"d", "e", "f"},
		'4': {"g", "h", "i"},
		'5': {"j", "k", "l"},
		'6': {"m", "n", "o"},
		'7': {"p", "q", "r", "s"},
		'8': {"t", "u", "v"},
		'9': {"w", "x", "y", "z"},
	}
	var res []string
	var str string
	l := len(digits)

	var dfs func(index int)
	dfs = func(index int) {
		if index >= l {
			return
		}
		if len(str) == l {
			res = append(res, str)
			return
		}

		// 先选择数字的映射
		arr := letterMap[digits[index]]
		for i := 0; i < len(arr); i++ {
			str += arr[i]
			// 选择下一个数字的映射
			dfs(index + 1)
			str = str[:len(str)-1]
		}
	}
	dfs(0)
	return res
}
