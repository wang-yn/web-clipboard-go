package models

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const OAuthHandoffCookieName = "oauth_handoff"

const (
	ClipboardExpirationUnitMinute = "minute"
	ClipboardExpirationUnitHour   = "hour"
	ClipboardExpirationUnitDay    = "day"
	ClipboardExpirationUnitNever  = "never"
)

// App represents the application state
type App struct {
	ClipboardData   map[string]*ClipboardItem
	DataMutex       *sync.RWMutex
	TempDir         string
	RateLimiter     RateLimiter
	Security        SecurityService
	CleanupTicker   *time.Ticker
	UserManager     UserManager
	AuthService     AuthService
	OAuthService    OAuthService
	SettingsService SettingsService
}

// ClipboardItem represents a clipboard entry (text or file)
type ClipboardItem struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	UserID      string    `json:"userId"`
	Content     string    `json:"content,omitempty"`
	FileName    string    `json:"fileName,omitempty"`
	FilePath    string    `json:"-"`
	ContentType string    `json:"contentType,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type SystemSettings struct {
	Auth      AuthSettings      `json:"auth"`
	Clipboard ClipboardSettings `json:"clipboard"`
}

type AuthSettings struct {
	PasswordLoginEnabled bool                `json:"passwordLoginEnabled"`
	OAuthAutoProvision   bool                `json:"oauthAutoProvision"`
	AllowedEmailDomains  []string            `json:"allowedEmailDomains"`
	Google               OAuthProviderConfig `json:"google"`
	GitHub               OAuthProviderConfig `json:"github"`
}

type OAuthProviderConfig struct {
	Enabled           bool   `json:"enabled"`
	ClientID          string `json:"clientId"`
	ClientSecret      string `json:"clientSecret,omitempty"`
	ClientSecretSet   bool   `json:"clientSecretSet,omitempty"`
	ClearClientSecret bool   `json:"clearClientSecret,omitempty"`
}

type ClipboardSettings struct {
	ExpirationValue int    `json:"expirationValue"`
	ExpirationUnit  string `json:"expirationUnit"`
}

type SystemSettingsResponse = SystemSettings

// User represents a user account
type User struct {
	ID         string             `json:"id"`
	Username   string             `json:"username"`
	Password   string             `json:"password,omitempty"` // bcrypt hashed; empty for external-only users
	Email      string             `json:"email"`
	Role       string             `json:"role"` // "admin" or "user"
	CreatedAt  time.Time          `json:"createdAt"`
	UpdatedAt  time.Time          `json:"updatedAt"`
	IsActive   bool               `json:"isActive"`
	Identities []ExternalIdentity `json:"identities,omitempty"`
}

// ExternalIdentity links a local user to a third-party identity provider.
type ExternalIdentity struct {
	Provider      string    `json:"provider"`
	Subject       string    `json:"subject"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"emailVerified"`
	Username      string    `json:"username"`
	DisplayName   string    `json:"displayName"`
	AvatarURL     string    `json:"avatarUrl"`
	LinkedAt      time.Time `json:"linkedAt"`
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

type RecentItemResponse struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	FileName    string    `json:"fileName,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type ListRecentItemsResponse struct {
	Items []RecentItemResponse `json:"items"`
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

type AuthProviderResponse struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type AuthProvidersResponse struct {
	Providers            []AuthProviderResponse `json:"providers"`
	PasswordLoginEnabled bool                   `json:"passwordLoginEnabled"`
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
	CreateExternalUser(identity ExternalIdentity, role string) (*User, error)
	LinkExternalIdentity(userID string, identity ExternalIdentity) (*User, error)
	GetUser(id string) *User
	GetUserByUsername(username string) *User
	GetUserByExternalIdentity(provider, subject string) *User
	GetUserByVerifiedEmail(email string) *User
	GetAllUsers() []User
	UpdateUser(id string, email, role string, isActive *bool) (*User, error)
	ChangePassword(id, newPassword string) error
	DeleteUser(id string) error
	ValidateCredentials(username, password string) (*User, error)
}

type SettingsService interface {
	GetSettings() SystemSettings
	GetSettingsResponse() SystemSettingsResponse
	SaveSettings(settings SystemSettings) error
}

type AuthService interface {
	CreateSession(userID string, rememberMe bool) (*Session, error)
	ValidateToken(token string) (*User, bool)
	GetUserByToken(token string) (*User, error)
	DeleteSession(token string)
	DeleteUserSessions(userID string)
	CleanupExpiredSessions()
}

type OAuthService interface {
	ListProviders() []AuthProviderResponse
	StartLogin(ctx context.Context, provider string) (string, error)
	HandleCallback(ctx context.Context, provider, code, state string) (*http.Cookie, error)
	CompleteLogin(handoff string) (LoginResponse, *http.Cookie, error)
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

func DefaultSystemSettings() SystemSettings {
	return SystemSettings{
		Auth: AuthSettings{
			PasswordLoginEnabled: true,
			AllowedEmailDomains:  []string{},
		},
		Clipboard: ClipboardSettings{
			ExpirationValue: 10,
			ExpirationUnit:  ClipboardExpirationUnitMinute,
		},
	}
}

func (settings ClipboardSettings) ExpiresAt(now time.Time) time.Time {
	value := settings.ExpirationValue
	if value <= 0 {
		value = 10
	}
	switch settings.ExpirationUnit {
	case ClipboardExpirationUnitNever:
		return time.Time{}
	case ClipboardExpirationUnitHour:
		return now.Add(time.Duration(value) * time.Hour)
	case ClipboardExpirationUnitDay:
		return now.Add(time.Duration(value) * 24 * time.Hour)
	default:
		return now.Add(time.Duration(value) * time.Minute)
	}
}

func ClipboardItemExpired(item *ClipboardItem, now time.Time) bool {
	if item == nil || item.ExpiresAt.IsZero() {
		return false
	}
	return item.ExpiresAt.Before(now)
}
