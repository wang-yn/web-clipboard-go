package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

type AuthService struct {
	sessions    map[string]*Session // key: token
	userManager *UserManager
	mutex       sync.RWMutex
}

func NewAuthService(userManager *UserManager) *AuthService {
	return &AuthService{
		sessions:    make(map[string]*Session),
		userManager: userManager,
	}
}

// CreateSession creates a new session for a user
func (as *AuthService) CreateSession(userID string, rememberMe bool) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	// Calculate expiration time
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().UTC().Add(7 * 24 * time.Hour) // 7 days
	} else {
		expiresAt = time.Now().UTC().Add(2 * time.Hour) // 2 hours
	}

	session := &Session{
		Token:      token,
		UserID:     userID,
		ExpiresAt:  expiresAt,
		RememberMe: rememberMe,
	}

	as.mutex.Lock()
	as.sessions[token] = session
	as.mutex.Unlock()

	return session, nil
}

// ValidateToken validates a token and returns whether it's valid
func (as *AuthService) ValidateToken(token string) (*User, bool) {
	as.mutex.RLock()
	session, exists := as.sessions[token]
	as.mutex.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if session is expired
	if session.ExpiresAt.Before(time.Now().UTC()) {
		as.DeleteSession(token)
		return nil, false
	}

	// Get user
	user := as.userManager.GetUser(session.UserID)
	if user == nil || !user.IsActive {
		as.DeleteSession(token)
		return nil, false
	}

	return user, true
}

// GetUserByToken gets a user by token
func (as *AuthService) GetUserByToken(token string) (*User, error) {
	user, valid := as.ValidateToken(token)
	if !valid {
		return nil, errors.New("invalid or expired token")
	}
	return user, nil
}

// DeleteSession deletes a session (logout)
func (as *AuthService) DeleteSession(token string) {
	as.mutex.Lock()
	delete(as.sessions, token)
	as.mutex.Unlock()
}

// DeleteUserSessions deletes all sessions for a user
func (as *AuthService) DeleteUserSessions(userID string) {
	as.mutex.Lock()
	defer as.mutex.Unlock()

	for token, session := range as.sessions {
		if session.UserID == userID {
			delete(as.sessions, token)
		}
	}
}

// CleanupExpiredSessions removes expired sessions
func (as *AuthService) CleanupExpiredSessions() {
	now := time.Now().UTC()
	as.mutex.Lock()
	defer as.mutex.Unlock()

	for token, session := range as.sessions {
		if session.ExpiresAt.Before(now) {
			delete(as.sessions, token)
		}
	}
}

// GetSessionCount returns the number of active sessions
func (as *AuthService) GetSessionCount() int {
	as.mutex.RLock()
	defer as.mutex.RUnlock()
	return len(as.sessions)
}

// generateToken generates a random token
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
