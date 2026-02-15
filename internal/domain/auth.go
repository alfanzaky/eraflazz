package domain

import (
	"strings"
	"time"
)

const (
	RoleReseller = "RESELLER"
	RoleAgent    = "AGENT"
	RoleMaster   = "MASTER"
	RoleAdmin    = "ADMIN"
	RoleH2H      = "H2H"
)

// AuthClaims represents validated JWT claims
type AuthClaims struct {
	UserID    string
	Role      string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// MapRoleToLevel converts role string to user level constant
func MapRoleToLevel(role string) int {
	switch strings.ToUpper(role) {
	case RoleAdmin:
		return LevelAdmin
	case RoleMaster:
		return LevelMaster
	case RoleAgent:
		return LevelAgent
	default:
		return LevelReseller
	}
}

// AuthService defines authentication helpers for JWT and H2H signature validation
type AuthService interface {
	GenerateAccessToken(user *User) (string, error)
	ValidateToken(token string) (*AuthClaims, error)
	ValidateH2HSignature(apiKey, signature, timestamp string, payload []byte) error
}

// MapLevelToRole converts user level to role string
func MapLevelToRole(level int) string {
	switch level {
	case LevelAdmin:
		return RoleAdmin
	case LevelMaster:
		return RoleMaster
	case LevelAgent:
		return RoleAgent
	case LevelReseller:
		return RoleReseller
	default:
		return RoleReseller
	}
}
