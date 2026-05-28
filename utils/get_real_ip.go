package utils

import (
	"net"
	"net/http"
	"strings"
)

func GetClientIP(r *http.Request) string {
	// 1. 优先尝试从 X-Forwarded-For 获取（如果是多级代理，取第一个 IP）
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For 的格式可能是: "client, proxy1, proxy2"
		ips := strings.Split(xForwardedFor, ",")
		realIP := strings.TrimSpace(ips[0])
		if realIP != "" {
			return realIP
		}
	}

	// 2. 其次尝试从 X-Real-IP 获取
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// 3. 如果都没有（比如本地直连测试），再降级使用 RemoteAddr
	// RemoteAddr 的格式通常是 "IP:Port"，需要切分出 IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // 切分失败则直接返回原字符串
	}
	return ip
}
