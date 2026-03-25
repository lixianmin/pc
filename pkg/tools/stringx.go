package tools

func StrTake(s string, n int) string {
	r := []rune(s)
	if len(r) < n {
		return s
	}

	return string(r[:n])
}
