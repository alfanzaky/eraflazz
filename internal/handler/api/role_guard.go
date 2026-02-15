package api

import (
	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
	"github.com/alfanzaky/eraflazz/pkg/xresponse"
	"github.com/gin-gonic/gin"
)

// RoleGuard provides helper functions for role-based access control in handlers
type RoleGuard struct{}

// NewRoleGuard creates a new role guard instance
func NewRoleGuard() *RoleGuard {
	return &RoleGuard{}
}

// GetCurrentUser extracts user information from context
func (rg *RoleGuard) GetCurrentUser(c *gin.Context) (userID, role string, userLevel int, exists bool) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return "", "", 0, false
	}

	roleVal, exists := c.Get("user_role")
	if !exists {
		return "", "", 0, false
	}

	levelVal, exists := c.Get("user_level")
	if !exists {
		return "", "", 0, false
	}

	userIDStr, ok := userIDVal.(string)
	if !ok {
		return "", "", 0, false
	}

	roleStr, ok := roleVal.(string)
	if !ok {
		return "", "", 0, false
	}

	levelInt, ok := levelVal.(int)
	if !ok {
		return "", "", 0, false
	}

	return userIDStr, roleStr, levelInt, true
}

// RequireRole checks if user has required role
func (rg *RoleGuard) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, role, _, exists := rg.GetCurrentUser(c)
		if !exists {
			logger.Warn("Access denied - user not authenticated",
				logger.String("required_role", requiredRole),
				logger.String("ip", c.ClientIP()),
			)
			xresponse.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		if role != requiredRole {
			logger.Warn("Access denied - insufficient role",
				logger.String("user_role", role),
				logger.String("required_role", requiredRole),
				logger.String("ip", c.ClientIP()),
			)
			xresponse.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		logger.Debug("Role access granted",
			logger.String("user_role", role),
			logger.String("required_role", requiredRole),
		)

		c.Next()
	}
}

// RequireMinimumLevel checks if user has minimum required level
func (rg *RoleGuard) RequireMinimumLevel(minLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, _, userLevel, exists := rg.GetCurrentUser(c)
		if !exists {
			logger.Warn("Access denied - user not authenticated",
				logger.String("required_level", string(rune(minLevel))),
				logger.String("ip", c.ClientIP()),
			)
			xresponse.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		if userLevel < minLevel {
			logger.Warn("Access denied - insufficient level",
				logger.String("user_level", string(rune(userLevel))),
				logger.String("required_level", string(rune(minLevel))),
				logger.String("ip", c.ClientIP()),
			)
			xresponse.Forbidden(c, "Insufficient permissions")
			c.Abort()
			return
		}

		logger.Debug("Level access granted",
			logger.String("user_level", string(rune(userLevel))),
			logger.String("required_level", string(rune(minLevel))),
		)

		c.Next()
	}
}

// RequireAdmin checks if user is admin
func (rg *RoleGuard) RequireAdmin() gin.HandlerFunc {
	return rg.RequireRole(domain.RoleAdmin)
}

// RequireMasterOrAbove checks if user is master or admin
func (rg *RoleGuard) RequireMasterOrAbove() gin.HandlerFunc {
	return rg.RequireMinimumLevel(domain.LevelMaster)
}

// RequireAgentOrAbove checks if user is agent, master, or admin
func (rg *RoleGuard) RequireAgentOrAbove() gin.HandlerFunc {
	return rg.RequireMinimumLevel(domain.LevelAgent)
}

// CanAccessOwnData checks if user can access their own data or if they have admin privileges
func (rg *RoleGuard) CanAccessOwnData(c *gin.Context, resourceUserID string) bool {
	userID, role, userLevel, exists := rg.GetCurrentUser(c)
	if !exists {
		return false
	}

	// Admin can access any data
	if role == domain.RoleAdmin || userLevel >= domain.LevelAdmin {
		logger.Debug("Admin access granted - can access any data",
			logger.String("admin_id", userID),
			logger.String("resource_user_id", resourceUserID),
		)
		return true
	}

	// Users can only access their own data
	if userID == resourceUserID {
		logger.Debug("Self access granted",
			logger.String("user_id", userID),
		)
		return true
	}

	logger.Warn("Access denied - cannot access other user's data",
		logger.String("user_id", userID),
		logger.String("resource_user_id", resourceUserID),
		logger.String("user_role", role),
	)

	return false
}

// CanAccessUplineData checks if user can access upline data (downline can see upline data for reporting)
func (rg *RoleGuard) CanAccessUplineData(c *gin.Context, resourceUserID string) bool {
	userID, role, userLevel, exists := rg.GetCurrentUser(c)
	if !exists {
		return false
	}

	// Admin can access any data
	if role == domain.RoleAdmin || userLevel >= domain.LevelAdmin {
		return true
	}

	// TODO: Implement proper upline checking from database
	// For now, only allow self-access
	return userID == resourceUserID
}

// IsH2HClient checks if current request is from H2H client
func (rg *RoleGuard) IsH2HClient(c *gin.Context) bool {
	_, exists := GetClientIDFromContext(c)
	return exists
}

// RequireUserOrH2H ensures request is from authenticated user or H2H client
func (rg *RoleGuard) RequireUserOrH2H() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if it's an H2H client
		if rg.IsH2HClient(c) {
			logger.Debug("H2H client access granted")
			c.Next()
			return
		}

		// Check if it's an authenticated user
		_, _, _, exists := rg.GetCurrentUser(c)
		if !exists {
			logger.Warn("Access denied - not authenticated user or H2H client",
				logger.String("ip", c.ClientIP()),
			)
			xresponse.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		logger.Debug("User access granted")
		c.Next()
	}
}

// LogAccess logs access with user information
func (rg *RoleGuard) LogAccess(c *gin.Context, action string, resource string) {
	if clientID, exists := GetClientIDFromContext(c); exists {
		logger.Info("H2H client action",
			logger.String("client_id", clientID),
			logger.String("action", action),
			logger.String("resource", resource),
			logger.String("ip", c.ClientIP()),
		)
		return
	}

	userID, role, _, exists := rg.GetCurrentUser(c)
	if exists {
		logger.Info("User action",
			logger.String("user_id", userID),
			logger.String("role", role),
			logger.String("action", action),
			logger.String("resource", resource),
			logger.String("ip", c.ClientIP()),
		)
	}
}
