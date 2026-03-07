package shared

import (
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

// RateLimiter IP başına sabit pencere sayacı ile istek sınırlandırır.
type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	limit   int
	window  time.Duration
}

// NewRateLimiter belirtilen limit ve zaman penceresiyle yeni bir RateLimiter oluşturur.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		entries: make(map[string]*rateLimitEntry),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

// cleanup süresi dolmuş kayıtları periyodik olarak temizler.
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window * 2)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, entry := range rl.entries {
			if now.Sub(entry.windowStart) > rl.window {
				delete(rl.entries, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow verilen IP için isteğe izin verilip verilmeyeceğini döner.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[ip]
	if !exists || now.Sub(entry.windowStart) > rl.window {
		rl.entries[ip] = &rateLimitEntry{count: 1, windowStart: now}
		return true
	}

	if entry.count >= rl.limit {
		return false
	}

	entry.count++
	return true
}

// Middleware rate limiting'i http.HandlerFunc olarak uygular.
func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !rl.Allow(ip) {
			http.Error(w, "Çok fazla istek gönderildi. Lütfen bir süre bekleyin.", http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}

// getClientIP isteğin kaynak IP adresini döner.
// Reverse proxy arkasında çalışıyorsa X-Forwarded-For header'ını kullanır.
func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
