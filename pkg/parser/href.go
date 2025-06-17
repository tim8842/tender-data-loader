package parser

import "strings"

func GetParamFromHref(href, param string) string {
	parts := strings.Split(href, "?")
	if len(parts) < 2 {
		return ""
	}
	for _, p := range strings.Split(parts[1], "&") {
		if strings.HasPrefix(p, param+"=") {
			return strings.TrimPrefix(p, param+"=")
		}
	}
	return ""
}
