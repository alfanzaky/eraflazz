package domain

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"` // Hidden in JSON
	FullName     *string   `json:"full_name" db:"full_name"`
	Phone        *string   `json:"phone" db:"phone"`
	
	// Hierarchy and permissions
	UplineID    *string `json:"upline_id" db:"upline_id"`
	Level       int     `json:"level" db:"level"`
	IsActive    bool    `json:"is_active" db:"is_active"`
	IsVerified  bool    `json:"is_verified" db:"is_verified"`
	
	// Financial information
	Balance         float64 `json:"balance" db:"balance"`
	CreditLimit     float64 `json:"credit_limit" db:"credit_limit"`
	MarkupPercentage float64 `json:"markup_percentage" db:"markup_percentage"`
	
	// Business settings
	AllowDebt           bool    `json:"allow_debt" db:"allow_debt"`
	MaxDailyTransaction float64 `json:"max_daily_transaction" db:"max_daily_transaction"`
	
	// Timestamps
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
}

// UserRepository defines operations for user data access
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByPhone(phone string) (*User, error)
	Update(user *User) error
	Delete(id string) error
	GetDownlines(uplineID string) ([]*User, error)
	UpdateBalance(id string, newBalance float64) error
	GetBalance(id string) (float64, error)
}

// UserUsecase defines business logic operations for users
type UserUsecase interface {
	Register(user *User) error
	Login(username, password string) (*User, error)
	UpdateProfile(id string, updates *User) error
	UpdateBalance(id string, amount float64, mutationType string, description string) error
	GetUserByID(id string) (*User, error)
	GetDownlines(uplineID string) ([]*User, error)
	DeactivateUser(id string) error
	VerifyUser(id string) error
}

// User validation rules
const (
	LevelReseller = 1
	LevelAgent    = 2
	LevelMaster   = 3
	LevelAdmin    = 4
)

// IsValidLevel checks if the user level is valid
func IsValidLevel(level int) bool {
	return level >= LevelReseller && level <= LevelAdmin
}

// CanHaveDownlines checks if user can have downlines based on level
func (u *User) CanHaveDownlines() bool {
	return u.Level >= LevelAgent
}

// GetEffectivePrice calculates the final price for a user based on their markup
func (u *User) GetEffectivePrice(basePrice float64) float64 {
	if u.Level == LevelAdmin {
		return basePrice // Admin gets base price
	}
	return basePrice * (1 + u.MarkupPercentage/100)
}

// HasSufficientBalance checks if user has enough balance for a transaction
func (u *User) HasSufficientBalance(amount float64) bool {
	if u.AllowDebt {
		return u.Balance+u.CreditLimit >= amount
	}
	return u.Balance >= amount
}
