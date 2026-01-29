package path

import "strings"

func IsPath(s string) bool {
	if s == "" {
		return false
	}
	if strings.Contains(s, "\x00") {
		return false
	}

	return isValidPath(s)
}
