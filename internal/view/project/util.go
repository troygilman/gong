package project

import (
	"strings"
)

func buildIncludeString(includes ...[]string) string {
	var result []string
	for _, include := range includes {
		result = append(result, include...)
	}
	for i, include := range result {
		result[i] = "#" + include
	}
	return strings.Join(result, ",")
}
