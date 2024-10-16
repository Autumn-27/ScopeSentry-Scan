// Package dirutils -----------------------------
// @file      : diff.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/5/7 21:38
// -------------------------------------------
package dirutils

type SequenceMatcher struct {
	a, b []rune
}

func NewSequenceMatcher(a, b string) *SequenceMatcher {
	return &SequenceMatcher{
		a: []rune(a),
		b: []rune(b),
	}
}

func (s *SequenceMatcher) Ratio() float64 {
	m := len(s.a)
	n := len(s.b)
	if m == 0 && n == 0 {
		return 1.0
	}

	// 仅用一个一维数组来保存前一行的数据
	curr := make([]int, n+1)

	for j := 0; j <= n; j++ {
		curr[j] = j
	}

	for i := 1; i <= m; i++ {
		prev := curr[0]
		curr[0] = i
		for j := 1; j <= n; j++ {
			temp := curr[j]
			if s.a[i-1] == s.b[j-1] {
				curr[j] = prev
			} else {
				curr[j] = 1 + dmin(curr[j], curr[j-1], prev)
			}
			prev = temp
		}
	}

	maxLen := dmax(m, n)
	return 1.0 - float64(curr[n])/float64(maxLen)
}

func (s *SequenceMatcher) Ratio2() float64 {
	m := len(s.a)
	n := len(s.b)
	if m == 0 && n == 0 {
		return 1.0
	}

	// 初始化一维数组
	curr := make([]int, n+1)

	for j := 0; j <= n; j++ {
		curr[j] = j
	}

	for i := 1; i <= m; i++ {
		prev := curr[0]
		curr[0] = i
		for j := 1; j <= n; j++ {
			temp := curr[j]
			cost := 0
			if s.a[i-1] != s.b[j-1] {
				cost = 1
			}
			curr[j] = dmin(curr[j]+1, // 插入
				prev+1,    // 删除
				temp+cost) // 替换
			prev = temp
		}
	}

	maxLen := dmax(m, n)
	return 1.0 - float64(curr[n])/float64(maxLen)
}

func dmin(a, b, c int) int {
	if a <= b && a <= c {
		return a
	} else if b <= a && b <= c {
		return b
	} else {
		return c
	}
}

func dmax(a, b int) int {
	if a >= b {
		return a
	} else {
		return b
	}
}
