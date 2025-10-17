package utils

import "net/url"

func IsUrl(str string) bool {
	// 使用url.ParseRequestURI解析URL
	parsedUrl, err := url.ParseRequestURI(str)
	if err != nil {
		return false // 解析失败，不是有效URL
	}

	// 验证URL必须包含 scheme (如http/https) 和 host (域名/IP)
	return parsedUrl.Scheme != "" && parsedUrl.Host != ""
}
