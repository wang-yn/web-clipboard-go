package services

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"web-clipboard-go/internal/models"
)

type SecurityService struct {
	failedAttempts     map[string]*models.FailedAttemptInfo
	blockedIPs         map[string]bool
	mutex              sync.RWMutex
	blockedExtensions  map[string]bool
	suspiciousPatterns []string
}

func NewSecurityService() *SecurityService {
	blockedExts := map[string]bool{
		".exe": true, ".bat": true, ".cmd": true, ".com": true, ".pif": true,
		".scr": true, ".vbs": true, ".js": true, ".jar": true, ".ps1": true,
		".sh": true, ".msi": true, ".dll": true, ".sys": true, ".php": true,
		".asp": true, ".aspx": true, ".jsp": true,
	}

	suspiciousPatterns := []string{
		"<script", "javascript:", "data:text/html", "eval(", "document.write",
		"base64,", "php://", "file://", "ftp://", "../../",
	}

	return &SecurityService{
		failedAttempts:     make(map[string]*models.FailedAttemptInfo),
		blockedIPs:         make(map[string]bool),
		blockedExtensions:  blockedExts,
		suspiciousPatterns: suspiciousPatterns,
	}
}

func (s *SecurityService) ValidateContentRequest(c interface{}, content string) bool {
	ip := s.GetClientIP(c)

	s.mutex.RLock()
	if s.blockedIPs[ip] {
		s.mutex.RUnlock()
		return false
	}
	s.mutex.RUnlock()

	if len(content) > 1024*1024 {
		s.recordFailedAttempt(ip, "Large content")
		return false
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range s.suspiciousPatterns {
		if strings.Contains(lowerContent, strings.ToLower(pattern)) {
			s.recordFailedAttempt(ip, fmt.Sprintf("Suspicious pattern: %s", pattern))
			return false
		}
	}

	return true
}

func (s *SecurityService) ValidateFileRequest(c interface{}) bool {
	ip := s.GetClientIP(c)

	s.mutex.RLock()
	blocked := s.blockedIPs[ip]
	s.mutex.RUnlock()

	return !blocked
}

func (s *SecurityService) ValidateFileType(fileName string) bool {
	if fileName == "" {
		return false
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	return !s.blockedExtensions[ext]
}

func (s *SecurityService) ValidateAccessRequest(c interface{}) bool {
	ip := s.GetClientIP(c)

	s.mutex.RLock()
	if s.blockedIPs[ip] {
		s.mutex.RUnlock()
		return false
	}

	if info, exists := s.failedAttempts[ip]; exists && info.Count > 50 {
		s.mutex.RUnlock()
		s.mutex.Lock()
		s.blockedIPs[ip] = true
		s.mutex.Unlock()
		log.Printf("Blocked IP %s for excessive failed attempts", ip)
		return false
	}
	s.mutex.RUnlock()

	return true
}

func (s *SecurityService) LogAccess(c interface{}, id, itemType string, success bool) {
	ip := s.GetClientIP(c)
	timestamp := time.Now().UTC()

	status := "SUCCESS"
	if !success {
		status = "FAILED"
		s.recordFailedAttempt(ip, fmt.Sprintf("Failed access to %s", id))
	}

	log.Printf("[%s] %s accessed %s %s: %s", timestamp.Format(time.RFC3339), ip, itemType, id, status)
}

func (s *SecurityService) GetClientIP(c interface{}) string {
	ctx, ok := c.(*gin.Context)
	if !ok {
		return ""
	}

	xForwardedFor := ctx.GetHeader("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	xRealIP := ctx.GetHeader("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	return ctx.ClientIP()
}

func (s *SecurityService) recordFailedAttempt(ip, reason string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if info, exists := s.failedAttempts[ip]; exists {
		info.Count++
		info.LastAttempt = time.Now().UTC()
		info.Reason = reason
	} else {
		s.failedAttempts[ip] = &models.FailedAttemptInfo{
			Count:       1,
			LastAttempt: time.Now().UTC(),
			Reason:      reason,
		}
	}

	if s.failedAttempts[ip].Count > 20 {
		s.blockedIPs[ip] = true
		log.Printf("Blocked IP %s after %d failed attempts. Latest: %s", ip, s.failedAttempts[ip].Count, reason)
	}
}

func (s *SecurityService) CleanupExpired() {
	cutoff := time.Now().UTC().Add(-1 * time.Hour)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	for ip, info := range s.failedAttempts {
		if info.LastAttempt.Before(cutoff) {
			delete(s.failedAttempts, ip)
		}
	}
}

type RateLimitService struct {
	ipLimits map[string]*models.RateLimitInfo
	mutex    sync.RWMutex
}

func NewRateLimitService() *RateLimitService {
	return &RateLimitService{
		ipLimits: make(map[string]*models.RateLimitInfo),
	}
}

func (r *RateLimitService) IsAllowed(ipAddress, endpoint string) bool {
	key := fmt.Sprintf("%s:%s", ipAddress, endpoint)
	now := time.Now().UTC()

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if info, exists := r.ipLimits[key]; exists {
		if now.Sub(info.WindowStart) > time.Minute {
			info.Count = 1
			info.WindowStart = now
		} else {
			info.Count++
		}
	} else {
		r.ipLimits[key] = &models.RateLimitInfo{
			Count:       1,
			WindowStart: now,
		}
	}

	limit := 50
	switch endpoint {
	case "POST":
		limit = 20
	case "GET":
		limit = 100
	}

	return r.ipLimits[key].Count <= limit
}

func (r *RateLimitService) CleanupExpired() {
	cutoff := time.Now().UTC().Add(-2 * time.Minute)

	r.mutex.Lock()
	defer r.mutex.Unlock()

	for key, info := range r.ipLimits {
		if info.WindowStart.Before(cutoff) {
			delete(r.ipLimits, key)
		}
	}
}
