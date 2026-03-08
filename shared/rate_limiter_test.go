package shared

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter_Allow_LimitAltinda(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Errorf("%d. istekte izin beklendi", i+1)
		}
	}
}

func TestRateLimiter_Allow_LimitAsildiktan_Sonra_Reddeder(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := 0; i < 3; i++ {
		rl.Allow("1.2.3.4")
	}

	if rl.Allow("1.2.3.4") {
		t.Error("4. istekte red beklendi")
	}
}

func TestRateLimiter_FarkliIP_BagimsizSayac(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	rl.Allow("1.1.1.1")
	if rl.Allow("1.1.1.1") {
		t.Error("2. istekte 1.1.1.1 için red beklendi")
	}
	if !rl.Allow("2.2.2.2") {
		t.Error("farklı IP için izin beklendi")
	}
}

func TestRateLimiter_PencereSifirlanir(t *testing.T) {
	rl := NewRateLimiter(2, 50*time.Millisecond)

	rl.Allow("5.5.5.5")
	rl.Allow("5.5.5.5")
	if rl.Allow("5.5.5.5") {
		t.Error("limitten sonra red beklendi")
	}

	time.Sleep(60 * time.Millisecond)

	if !rl.Allow("5.5.5.5") {
		t.Error("pencere sıfırlandıktan sonra izin beklendi")
	}
}

func TestRateLimiter_Middleware_LimitAltinda_200(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	handler := rl.Middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		w := httptest.NewRecorder()
		handler(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("%d. istek: 200 beklendi, alınan=%d", i+1, w.Code)
		}
	}
}

func TestRateLimiter_Middleware_LimitAsilinca_429(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)
	handler := rl.Middleware(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.2:5678"
		w := httptest.NewRecorder()
		handler(w, req)
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.2:5678"
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("429 beklendi, alınan=%d", w.Code)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 10.0.0.1")

	ip := getClientIP(req)
	if ip != "203.0.113.1" {
		t.Errorf("IP=203.0.113.1 beklendi, alınan=%q", ip)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.5:4321"

	ip := getClientIP(req)
	if ip != "192.168.1.5" {
		t.Errorf("IP=192.168.1.5 beklendi, alınan=%q", ip)
	}
}

func TestRateLimiter_Allow_Limitten_Sonra_IzinYok(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	ip := "7.7.7.7"

	for i := 0; i < 5; i++ {
		rl.Allow(ip)
	}

	for i := 0; i < 3; i++ {
		if rl.Allow(ip) {
			t.Errorf("limit sonrası %d. istekte red beklendi", i+1)
		}
	}
}
