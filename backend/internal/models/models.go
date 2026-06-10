package models

import (
	"sync"
	"time"
)

// App represents the application state
type App struct {
	ClipboardData map[string]*ClipboardItem
	DataMutex     *sync.RWMutex
	TempDir       string
	RateLimiter   RateLimiter
	Security      SecurityService
	CleanupTicker *time.Ticker
	UserManager   UserManager
	AuthService   AuthService
}

// ClipboardItem represents a clipboard entry (text or file)
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

// UsersData represents the structure of users.json file
type UsersData struct {
	Users []User `json:"users"`
}

// Request/Response types for clipboard operations
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

// Request/Response types for authentication
type LoginRequest struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"rememberMe"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

// Request types for user management
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Role     string `json:"role"` // optional, defaults to "user"
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive *bool  `json:"isActive"` // pointer to distinguish false from omitted
}

type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required"`
}

// Service interfaces - these are kept as interface{} for flexibility
// but in practice they expect *gin.Context
type UserManager interface {
	CreateUser(username, password, email, role string) (*User, error)
	GetUser(id string) *User
	GetUserByUsername(username string) *User
	GetAllUsers() []User
	UpdateUser(id string, email, role string, isActive *bool) (*User, error)
	ChangePassword(id, newPassword string) error
	DeleteUser(id string) error
	ValidateCredentials(username, password string) (*User, error)
}

type AuthService interface {
	CreateSession(userID string, rememberMe bool) (*Session, error)
	ValidateToken(token string) (*User, bool)
	GetUserByToken(token string) (*User, error)
	DeleteSession(token string)
	DeleteUserSessions(userID string)
	CleanupExpiredSessions()
}

type SecurityService interface {
	ValidateContentRequest(c interface{}, content string) bool
	ValidateFileRequest(c interface{}) bool
	ValidateFileType(fileName string) bool
	ValidateAccessRequest(c interface{}) bool
	LogAccess(c interface{}, id, itemType string, success bool)
	CleanupExpired()
	GetClientIP(c interface{}) string
}

type RateLimiter interface {
	IsAllowed(ipAddress, endpoint string) bool
	CleanupExpired()
}

// FailedAttemptInfo tracks failed access attempts
type FailedAttemptInfo struct {
	Count       int
	LastAttempt time.Time
	Reason      string
}

// RateLimitInfo tracks rate limit info per IP
type RateLimitInfo struct {
	Count       int
	WindowStart time.Time
}

// ToUserResponse converts User to UserResponse (without password)
func ToUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		IsActive:  user.IsActive,
	}
}
