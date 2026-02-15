package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/alfanzaky/eraflazz/config"
	"github.com/alfanzaky/eraflazz/internal/domain"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token expired")
	ErrSignatureInvalid = errors.New("invalid signature")
)

type customClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuthService implements domain.AuthService using JWT + HMAC signature for H2H
type JWTAuthService struct {
	cfg config.AuthConfig
}

// NewJWTAuthService creates a new auth service instance
func NewJWTAuthService(cfg config.AuthConfig) *JWTAuthService {
	return &JWTAuthService{cfg: cfg}
}

func (s *JWTAuthService) accessTTL() time.Duration {
	if s.cfg.AccessTokenTTL <= 0 {
		return 24 * time.Hour
	}
	return s.cfg.AccessTokenTTL
}

// GenerateAccessToken creates signed JWT access token for the given user
func (s *JWTAuthService) GenerateAccessToken(user *domain.User) (string, error) {
	if user == nil || user.ID == "" {
		return "", fmt.Errorf("invalid user payload")
	}

	now := time.Now()
	claims := &customClaims{
		Role: domain.MapLevelToRole(user.Level),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL())),
			ID:        fmt.Sprintf("%s-%d", user.ID, now.UnixNano()),
		},
	}
	if audience := strings.TrimSpace(s.cfg.Audience); audience != "" {
		claims.Audience = jwt.ClaimStrings{audience}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.cfg.AccessSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signed, nil
}

// ValidateToken parses and validates JWT token and returns AuthClaims
func (s *JWTAuthService) ValidateToken(token string) (*domain.AuthClaims, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	claims := &customClaims{}
	options := []jwt.ParserOption{jwt.WithIssuedAt(), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name})}
	if iss := strings.TrimSpace(s.cfg.Issuer); iss != "" {
		options = append(options, jwt.WithIssuer(iss))
	}
	if aud := strings.TrimSpace(s.cfg.Audience); aud != "" {
		options = append(options, jwt.WithAudience(aud))
	}

	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.AccessSecret), nil
	}, options...)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !parsed.Valid {
		return nil, ErrInvalidToken
	}

	role := strings.ToUpper(claims.Role)
	if role == "" {
		role = domain.RoleReseller
	}

	return &domain.AuthClaims{
		UserID:    claims.Subject,
		Role:      role,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}

// ValidateH2HSignature validates H2H signature using configured secret
func (s *JWTAuthService) ValidateH2HSignature(apiKey, signature, timestamp string, payload []byte) error {
	if s.cfg.H2HAPIKey == "" || s.cfg.H2HAPISecret == "" {
		return fmt.Errorf("H2H credentials not configured")
	}
	if apiKey != s.cfg.H2HAPIKey {
		return ErrSignatureInvalid
	}
	if signature == "" || timestamp == "" {
		return ErrSignatureInvalid
	}

	mac := hmac.New(sha256.New, []byte(s.cfg.H2HAPISecret))
	mac.Write([]byte(timestamp))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(strings.ToLower(signature)), []byte(strings.ToLower(expected))) {
		return ErrSignatureInvalid
	}

	return nil
}
