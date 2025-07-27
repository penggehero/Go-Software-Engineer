package code

/*
5. 最长回文子串
给你一个字符串 s，找到 s 中最长的 回文 子串。

示例 1：
输入：s = "babad"
输出："bab"
解释："aba" 同样是符合题意的答案。

示例 2：
输入：s = "cbbd"
输出："bb"
*/

// longestPalindrome 最长回文子串
func longestPalindrome(s string) string {
	n := len(s)
	if n < 2 {
		return s
	}

	maxLen := 1
	begin := 0

	// dp[i][j] 表示 s[i..j] 是否是回文串
	dp := make([][]bool, n)
	for i := range dp {
		dp[i] = make([]bool, n)
		dp[i][i] = true // 所有长度为1的子串都是回文串
	}

	// 枚举子串长度
	for L := 2; L <= n; L++ {
		for i := 0; i < n; i++ {
			j := L + i - 1
			// 出界，直接退出
			if j >= n {
				break
			}
			if s[i] != s[j] {
				dp[i][j] = false
			} else {
				// 组成 aa 或者 aba 的情况，默认为回文
				if j-i < 3 {
					dp[i][j] = true
				} else {
					// 由dp[i+1][j-1] 决定是否回文
					dp[i][j] = dp[i+1][j-1]
				}
			}
			// 如果dp[i][j] 回文，且长度比maxLen大，记录长度和begin
			if dp[i][j] && j-i+1 > maxLen {
				maxLen = j - i + 1
				begin = i
			}
		}
	}
	return s[begin : begin+maxLen]
}
