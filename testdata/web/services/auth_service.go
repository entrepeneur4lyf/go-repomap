package services

import (
	"sync"
	"time"
)

type TokenInfo struct {
	UserID    string
	ExpiresAt time.Time
}

type AuthService struct {
	tokens map[string]TokenInfo // token -> token info
	mu     sync.RWMutex
}

func NewAuthService() *AuthService {
	return &AuthService{
		tokens: make(map[string]TokenInfo),
	}
}

func (s *AuthService) ValidateToken(token string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.tokens[token]
	if !exists {
		return false
	}

	// Check if token has expired
	if time.Now().After(info.ExpiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.tokens, token)
		s.mu.Unlock()
		s.mu.RLock()
		return false
	}

	return true
}

func (s *AuthService) CreateToken(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// In a real application, we would:
	// 1. Generate a proper JWT or other secure token
	// 2. Sign it with a private key
	// 3. Include proper claims and headers
	// 4. Handle errors properly

	// This is just a simple example
	token := "token_" + userID + "_" + time.Now().Format(time.RFC3339)
	s.tokens[token] = TokenInfo{
		UserID:    userID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return token
}

func (s *AuthService) RevokeToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tokens, token)
}

func (s *AuthService) GetUserID(token string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.tokens[token]
	if !exists || time.Now().After(info.ExpiresAt) {
		return "", false
	}

	return info.UserID, true
}

func (s *AuthService) CleanupExpiredTokens() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for token, info := range s.tokens {
		if now.After(info.ExpiresAt) {
			delete(s.tokens, token)
		}
	}
}
