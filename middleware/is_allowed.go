package middleware

import (
	"log"
	"net/http"
	"omiro/helper"
	"omiro/redis"
	"time"
)

func AllowHandshake(ip string, limit int) (bool, error) {
	allowed, err := redis.CheckRateLimit(ip, limit, 1*time.Minute)
	if err != nil {
		return false, err
	}
	return allowed, nil
}

func EnsureUpgradeChecks(w http.ResponseWriter, r *http.Request) bool {

	origin := r.Header.Get("Origin")
	// TODO: Add allowed origins
	// For development, allow localhost and ngrok
	allowedOrigins := []string{
		"http://localhost:8080",
		"https://205a7133ed0f.ngrok-free.app",
	}

	// If no origin header (e.g., from same-origin), allow it
	if origin != "" {
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}
		if !allowed {
			log.Printf("Forbidden origin: %s", origin)
			http.Error(w, "Forbidden origin", http.StatusForbidden)
			return false
		}
	}

	ip := helper.GetRealIP(r)
	log.Printf("IP: %s", ip)
	ok, err := AllowHandshake(ip, 60)
	if err != nil || !ok {
		http.Error(w, "Too many requests", http.StatusTooManyRequests)
		return false
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		token = r.Header.Get("X-Session-Token")
	}
	log.Printf("Token: %s", token)
	if !VerifySessionToken(token) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	log.Printf("Token verified")
	return true
}
