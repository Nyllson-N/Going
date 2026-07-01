package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// visitor controla o número de requisições de um cliente dentro da janela atual.
type visitor struct {
	tokens   int
	windowAt time.Time
}

// RateLimiter limita as requisições por IP: até `limit` requisições a cada `period`.
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
	limit    int
	period   time.Duration
}

// NewRateLimiter cria um limitador para `limit` requisições a cada `period`.
func NewRateLimiter(limit int, period time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    limit,
		period:   period,
	}
}

// allow informa se o IP pode fazer mais uma requisição na janela atual.
func (rl *RateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, ok := rl.visitors[ip]
	if !ok || now.Sub(v.windowAt) >= rl.period {
		// Nova janela de tempo para o cliente.
		rl.visitors[ip] = &visitor{tokens: rl.limit - 1, windowAt: now}
		return true
	}

	if v.tokens <= 0 {
		return false
	}
	v.tokens--
	return true
}

// Handler é o middleware HTTP que aplica o rate limit.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.allow(clientIP(r)) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP extrai o IP do cliente a partir da requisição.
func clientIP(r *http.Request) string {
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}
	return r.RemoteAddr
}
