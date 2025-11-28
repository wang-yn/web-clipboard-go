package main

import (
	"sync"
	"time"
)

type App struct {
	clipboardData map[string]*ClipboardItem
	dataMutex     *sync.RWMutex
	tempDir       string
	rateLimiter   *RateLimitService
	security      *SecurityService
	cleanupTicker *time.Ticker
	userManager   *UserManager
	authService   *AuthService
}

type ClipboardItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Content     string    `json:"content,omitempty"`
	FileName    string    `json:"fileName,omitempty"`
	FilePath    string    `json:"-"`
	ContentType string    `json:"contentType,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type TextRequest struct {
	Content string `json:"content" binding:"required"`
}

type SaveTextResponse struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type GetTextResponse struct {
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

type SaveFileResponse struct {
	ID        string    `json:"id"`
	FileName  string    `json:"fileName"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type CleanupResponse struct {
	RemovedCount int `json:"removedCount"`
}

// User represents a user account
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password"` // bcrypt hashed
	Email     string    `json:"email"`
	Role      string    `json:"role"` // "admin" or "user"
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsActive  bool      `json:"isActive"`
}

// UserResponse is the user data returned to clients (without password)
type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsActive  bool      `json:"isActive"`
}

// Session represents a user session
type Session struct {
	Token      string    `json:"token"`
	UserID     string    `json:"userId"`
	ExpiresAt  time.Time `json:"expiresAt"`
	RememberMe bool      `json:"rememberMe"`
}

// LoginRequest is the login request payload
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`
}

// LoginResponse is the login response payload
type LoginResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

// CreateUserRequest is the request to create a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Role     string `json:"role"` // optional, defaults to "user"
}

// UpdateUserRequest is the request to update a user
type UpdateUserRequest struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive *bool  `json:"isActive"` // pointer to distinguish false from omitted
}

// ChangePasswordRequest is the request to change user password
type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

// UsersData represents the structure of users.json file
type UsersData struct {
	Users []User `json:"users"`
}
