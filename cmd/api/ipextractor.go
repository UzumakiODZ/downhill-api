package cmd

import (
	"net"
	"net/http"
	"strings"
)

const (
	trueClientIP  = "True-Client-Ip"
	xForwardedFor = "X-Forwarded-For"
)

// GetRemoteAddr returns the original client's IP address, optimized for Render.
func GetRemoteAddr(r *http.Request) string {
	// 1. Try Render's explicit True-Client-Ip header first
	if clientIP := r.Header.Get(trueClientIP); clientIP != "" {
		return strings.TrimSpace(clientIP)
	}

	// 2. Fallback to X-Forwarded-For (Render ensures the first IP is the real client)
	if xff := r.Header.Get(xForwardedFor); xff != "" {
		if parts := strings.Split(xff, ","); len(parts) > 0 {
			clientIP := strings.TrimSpace(parts[0])
			if clientIP != "" {
				return clientIP
			}
		}
	}

	// 3. Fallback to r.RemoteAddr for local development (stripping the port)
	remoteAddr := r.RemoteAddr
	if ip, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return ip
	}

	return remoteAddr
}