package helper

import (
	"net"
	"net/http"
	"strings"
)

func GetRealIP(r *http.Request) string {

	if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
		return cfIP
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.TrimSpace(strings.Split(xff, ",")[0])
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
