package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// createUser creates a new user (admin only)
func (app *App) createUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Create user
	user, err := app.userManager.CreateUser(req.Username, req.Password, req.Email, req.Role)
	if err != nil {
		log.Printf("Failed to create user '%s': %v", req.Username, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUser := c.MustGet("user").(*User)
	log.Printf("User '%s' created by admin '%s'", user.Username, currentUser.Username)

	c.JSON(http.StatusCreated, ToUserResponse(user))
}

// listUsers returns all users (admin only)
func (app *App) listUsers(c *gin.Context) {
	users := app.userManager.GetAllUsers()

	// Convert to response format (without passwords)
	responses := make([]UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": responses,
		"count": len(responses),
	})
}

// getUser returns a single user's details
func (app *App) getUser(c *gin.Context) {
	id := c.Param("id")

	user := app.userManager.GetUser(id)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check permissions: users can view their own info, admins can view any user
	currentUser := c.MustGet("user").(*User)
	if currentUser.Role != "admin" && currentUser.ID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	c.JSON(http.StatusOK, ToUserResponse(user))
}

// updateUser updates a user's information
func (app *App) updateUser(c *gin.Context) {
	id := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check permissions: users can update their own email, admins can update any user
	currentUser := c.MustGet("user").(*User)
	if currentUser.Role != "admin" && currentUser.ID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Non-admins cannot change role or isActive
	if currentUser.Role != "admin" {
		if req.Role != "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can change roles"})
			return
		}
		if req.IsActive != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can change active status"})
			return
		}
	}

	// Update user
	user, err := app.userManager.UpdateUser(id, req.Email, req.Role, req.IsActive)
	if err != nil {
		log.Printf("Failed to update user '%s': %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("User '%s' updated by '%s'", user.Username, currentUser.Username)

	// If user is deactivated, delete all their sessions
	if req.IsActive != nil && !*req.IsActive {
		app.authService.DeleteUserSessions(id)
		log.Printf("All sessions deleted for deactivated user '%s'", user.Username)
	}

	c.JSON(http.StatusOK, ToUserResponse(user))
}

// deleteUser deletes a user
func (app *App) deleteUser(c *gin.Context) {
	id := c.Param("id")

	currentUser := c.MustGet("user").(*User)

	// Prevent users from deleting themselves
	if currentUser.ID == id {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}

	// Get user info before deletion for logging
	user := app.userManager.GetUser(id)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Delete user
	if err := app.userManager.DeleteUser(id); err != nil {
		log.Printf("Failed to delete user '%s': %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete all sessions for this user
	app.authService.DeleteUserSessions(id)

	log.Printf("User '%s' deleted by admin '%s'", user.Username, currentUser.Username)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// changeUserPassword changes a user's password
func (app *App) changeUserPassword(c *gin.Context) {
	id := c.Param("id")

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Check permissions: users can change their own password, admins can change any password
	currentUser := c.MustGet("user").(*User)
	if currentUser.Role != "admin" && currentUser.ID != id {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Change password
	if err := app.userManager.ChangePassword(id, req.NewPassword); err != nil {
		log.Printf("Failed to change password for user '%s': %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := app.userManager.GetUser(id)
	log.Printf("Password changed for user '%s' by '%s'", user.Username, currentUser.Username)

	// Delete all sessions for this user (force re-login)
	app.authService.DeleteUserSessions(id)

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully. Please login again."})
}
