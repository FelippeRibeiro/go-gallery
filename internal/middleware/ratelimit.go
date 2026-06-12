package middleware

import (
	"sync"
	"time"
)

// Rate limit de login em memória: até loginMaxAttempts tentativas falhas por
// chave (IP+email) dentro de loginWindow. Suficiente para instância única —
// com múltiplas réplicas, mover para um armazenamento compartilhado (Redis).

const (
	loginMaxAttempts = 10
	loginWindow      = 15 * time.Minute
)

var loginLimiter = struct {
	mu        sync.Mutex
	attempts  map[string][]time.Time
	lastSweep time.Time
}{attempts: make(map[string][]time.Time), lastSweep: time.Now()}

// LoginAllowed registra uma tentativa de login para a chave e informa se ela
// ainda está dentro do limite. Tentativas bem-sucedidas devem ser zeradas com
// LoginSucceeded para não contarem contra o usuário.
func LoginAllowed(key string) bool {
	now := time.Now()
	corte := now.Add(-loginWindow)

	loginLimiter.mu.Lock()
	defer loginLimiter.mu.Unlock()

	// Varredura periódica para o mapa não crescer sem limite.
	if now.Sub(loginLimiter.lastSweep) > loginWindow {
		for k, ts := range loginLimiter.attempts {
			if len(ts) == 0 || ts[len(ts)-1].Before(corte) {
				delete(loginLimiter.attempts, k)
			}
		}
		loginLimiter.lastSweep = now
	}

	validas := loginLimiter.attempts[key][:0]
	for _, t := range loginLimiter.attempts[key] {
		if t.After(corte) {
			validas = append(validas, t)
		}
	}
	if len(validas) >= loginMaxAttempts {
		loginLimiter.attempts[key] = validas
		return false
	}
	loginLimiter.attempts[key] = append(validas, now)
	return true
}

// LoginSucceeded limpa o contador da chave após um login correto.
func LoginSucceeded(key string) {
	loginLimiter.mu.Lock()
	delete(loginLimiter.attempts, key)
	loginLimiter.mu.Unlock()
}
