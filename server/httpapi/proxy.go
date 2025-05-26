package httpapi

import (
	"strings"

	"github.com/WindowsSov8forUs/go-kyutorin/processor"
)

// couldBeProxied 是否可代理
func couldBeProxied(url string) bool {
	for _, proxyUrl := range processor.ProxyUrls() {
		if strings.HasPrefix(url, proxyUrl) {
			return true
		}
	}
	return false
}
