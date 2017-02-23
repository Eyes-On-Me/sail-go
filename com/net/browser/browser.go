package browser

import (
	"strings"
)

// 是否移动端
func IsMobile(user_agent string) bool {
	user_agent = strings.ToLower(user_agent)
	arr := [...]string{"iphone", "android", "phone", "mobile"}
	for _, str := range arr {
		if strings.Index(user_agent, str) != -1 {
			return true
		}
	}
	return false
}
