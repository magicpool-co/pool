package common

import (
	"strings"
)

func StringsEqualInsensitive(a, b string) bool {
	return strings.ToLower(a) == strings.ToLower(b)
}
