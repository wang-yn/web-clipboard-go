package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"web-clipboard-go/internal/models"
	"web-clipboard-go/internal/utils"
)

type UserManager struct {
	users    map[string]*models.User // key: user ID
	filePath string
	mutex    sync.RWMutex
}

func NewUserManager(dataDir string) (*UserManager, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	um := &UserManager{
		users:    make(map[string]*models.User),
		filePath: filepath.Join(dataDir, "users.json"),
	}

	// Load existing users or create default admin
	if err := um.loadUsers(); err != nil {
		return nil, err
	}

	// Create default admin if no users exist
	if len(um.users) == 0 {
		if err := um.createDefaultAdmin(); err != nil {
			return nil, err
		}
	}

	return um, nil
}

// loadUsers loads users from JSON file
func (um *UserManager) loadUsers() error {
	data, err := os.ReadFile(um.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, start with empty users
			return nil
		}
		return fmt.Errorf("failed to read users file: %w", err)
	}

	var usersData models.UsersData
	if err := json.Unmarshal(data, &usersData); err != nil {
		return fmt.Errorf("failed to parse users file: %w", err)
	}

	um.mutex.Lock()
	defer um.mutex.Unlock()

	for i := range usersData.Users {
		user := &usersData.Users[i]
		um.users[user.ID] = user
	}

	return nil
}

// saveUsers saves users to JSON file
func (um *UserManager) saveUsers() error {
	um.mutex.RLock()
	usersList := make([]models.User, 0, len(um.users))
	for _, user := range um.users {
		usersList = append(usersList, *user)
	}
	um.mutex.RUnlock()

	usersData := models.UsersData{Users: usersList}
	data, err := json.MarshalIndent(usersData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}

	if err := os.WriteFile(um.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write users file: %w", err)
	}

	return nil
}

// createDefaultAdmin creates the default admin account
func (um *UserManager) createDefaultAdmin() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	admin := &models.User{
		ID:        utils.GenerateUUID(),
		Username:  "admin",
		Password:  string(hashedPassword),
		Email:     "admin@localhost",
		Role:      "admin",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		IsActive:  true,
	}

	um.mutex.Lock()
	um.users[admin.ID] = admin
	um.mutex.Unlock()

	if err := um.saveUsers(); err != nil {
		return err
	}

	fmt.Println("===========================================")
	fmt.Println("Default admin account created:")
	fmt.Println("  Username: admin")
	fmt.Println("  Password: admin123")
	fmt.Println("  Please change the password after first login!")
	fmt.Println("===========================================")

	return nil
}

// CreateUser creates a new user
func (um *UserManager) CreateUser(username, password, email, role string) (*models.User, error) {
	// Validate inputs
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if password == "" {
		return nil, errors.New("password cannot be empty")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}
	email = strings.TrimSpace(email)
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}

	// Default role to "user" if not specified
	if role == "" {
		role = "user"
	}
	if role != "admin" && role != "user" {
		return nil, errors.New("role must be 'admin' or 'user'")
	}

	// Check if username already exists
	if um.GetUserByUsername(username) != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:        utils.GenerateUUID(),
		Username:  username,
		Password:  string(hashedPassword),
		Email:     email,
		Role:      role,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		IsActive:  true,
	}

	um.mutex.Lock()
	um.users[user.ID] = user
	um.mutex.Unlock()

	if err := um.saveUsers(); err != nil {
		// Rollback
		um.mutex.Lock()
		delete(um.users, user.ID)
		um.mutex.Unlock()
		return nil, err
	}

	return user, nil
}

// GetUser gets a user by ID
func (um *UserManager) GetUser(id string) *models.User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()
	return um.users[id]
}

// GetUserByUsername gets a user by username
func (um *UserManager) GetUserByUsername(username string) *models.User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	username = strings.ToLower(strings.TrimSpace(username))
	for _, user := range um.users {
		if strings.ToLower(user.Username) == username {
			return user
		}
	}
	return nil
}

// GetAllUsers returns all users
func (um *UserManager) GetAllUsers() []models.User {
	um.mutex.RLock()
	defer um.mutex.RUnlock()

	users := make([]models.User, 0, len(um.users))
	for _, user := range um.users {
		users = append(users, *user)
	}
	return users
}

// UpdateUser updates a user's information
func (um *UserManager) UpdateUser(id string, email, role string, isActive *bool) (*models.User, error) {
	um.mutex.Lock()
	user, exists := um.users[id]
	if !exists {
		um.mutex.Unlock()
		return nil, errors.New("user not found")
	}

	// Update fields if provided
	if email != "" {
		user.Email = strings.TrimSpace(email)
	}
	if role != "" {
		if role != "admin" && role != "user" {
			um.mutex.Unlock()
			return nil, errors.New("role must be 'admin' or 'user'")
		}
		user.Role = role
	}
	if isActive != nil {
		user.IsActive = *isActive
	}

	user.UpdatedAt = time.Now().UTC()
	um.mutex.Unlock()

	if err := um.saveUsers(); err != nil {
		return nil, err
	}

	return user, nil
}

// ChangePassword changes a user's password
func (um *UserManager) ChangePassword(id, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	um.mutex.Lock()
	user, exists := um.users[id]
	if !exists {
		um.mutex.Unlock()
		return errors.New("user not found")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		um.mutex.Unlock()
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now().UTC()
	um.mutex.Unlock()

	return um.saveUsers()
}

// DeleteUser deletes a user
func (um *UserManager) DeleteUser(id string) error {
	um.mutex.Lock()
	user, exists := um.users[id]
	if !exists {
		um.mutex.Unlock()
		return errors.New("user not found")
	}

	// Prevent deleting the last admin
	if user.Role == "admin" {
		adminCount := 0
		for _, u := range um.users {
			if u.Role == "admin" && u.IsActive {
				adminCount++
			}
		}
		if adminCount <= 1 {
			um.mutex.Unlock()
			return errors.New("cannot delete the last admin user")
		}
	}

	delete(um.users, id)
	um.mutex.Unlock()

	return um.saveUsers()
}

// ValidateCredentials validates username and password
func (um *UserManager) ValidateCredentials(username, password string) (*models.User, error) {
	user := um.GetUserByUsername(username)
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}
