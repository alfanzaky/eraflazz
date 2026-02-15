package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
)

type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, full_name, phone, 
			upline_id, level, is_active, is_verified, balance, credit_limit, 
			markup_percentage, allow_debt, max_daily_transaction)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err := r.db.Exec(query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.FullName, user.Phone, user.UplineID, user.Level,
		user.IsActive, user.IsVerified, user.Balance, user.CreditLimit,
		user.MarkupPercentage, user.AllowDebt, user.MaxDailyTransaction,
	)

	if err != nil {
		logger.Error("Failed to create user", 
			logger.String("username", user.Username),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User created successfully", 
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
	)

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE id = $1
	`

	var user domain.User
	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by ID", 
			logger.String("user_id", id),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(username string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE username = $1
	`

	var user domain.User
	err := r.db.Get(&user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by username", 
			logger.String("username", username),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`

	var user domain.User
	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by email", 
			logger.String("email", email),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByPhone retrieves a user by phone number
func (r *userRepository) GetByPhone(phone string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE phone = $1
	`

	var user domain.User
	err := r.db.Get(&user, query, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get user by phone", 
			logger.String("phone", phone),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(user *domain.User) error {
	query := `
		UPDATE users SET 
			username = $2, email = $3, password_hash = $4, full_name = $5, phone = $6,
			upline_id = $7, level = $8, is_active = $9, is_verified = $10,
			balance = $11, credit_limit = $12, markup_percentage = $13,
			allow_debt = $14, max_daily_transaction = $15, last_login_at = $16
		WHERE id = $1
	`

	result, err := r.db.Exec(query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.FullName, user.Phone, user.UplineID, user.Level,
		user.IsActive, user.IsVerified, user.Balance, user.CreditLimit,
		user.MarkupPercentage, user.AllowDebt, user.MaxDailyTransaction,
		user.LastLoginAt,
	)

	if err != nil {
		logger.Error("Failed to update user", 
			logger.String("user_id", user.ID),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Info("User updated successfully", 
		logger.String("user_id", user.ID),
		logger.String("username", user.Username),
	)

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(id string) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		logger.Error("Failed to delete user", 
			logger.String("user_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Info("User deleted successfully", 
		logger.String("user_id", id),
	)

	return nil
}

// GetDownlines retrieves all downlines of a user
func (r *userRepository) GetDownlines(uplineID string) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE upline_id = $1 ORDER BY created_at DESC
	`

	var users []*domain.User
	err := r.db.Select(&users, query, uplineID)
	if err != nil {
		logger.Error("Failed to get downlines", 
			logger.String("upline_id", uplineID),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get downlines: %w", err)
	}

	return users, nil
}

// UpdateBalance updates user balance
func (r *userRepository) UpdateBalance(id string, newBalance float64) error {
	query := `UPDATE users SET balance = $2 WHERE id = $1`

	result, err := r.db.Exec(query, id, newBalance)
	if err != nil {
		logger.Error("Failed to update balance", 
			logger.String("user_id", id),
			logger.Float64("new_balance", newBalance),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	logger.Info("Balance updated successfully", 
		logger.String("user_id", id),
		logger.Float64("new_balance", newBalance),
	)

	return nil
}

// GetBalance retrieves user balance
func (r *userRepository) GetBalance(id string) (float64, error) {
	query := `SELECT balance FROM users WHERE id = $1`

	var balance float64
	err := r.db.Get(&balance, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		logger.Error("Failed to get balance", 
			logger.String("user_id", id),
			logger.ErrorField(err),
		)
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// UpdateLastLogin updates user's last login time
func (r *userRepository) UpdateLastLogin(id string) error {
	query := `UPDATE users SET last_login_at = $2 WHERE id = $1`
	now := time.Now()

	result, err := r.db.Exec(query, id, now)
	if err != nil {
		logger.Error("Failed to update last login", 
			logger.String("user_id", id),
			logger.ErrorField(err),
		)
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetActiveUsers retrieves all active users
func (r *userRepository) GetActiveUsers() ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE is_active = true ORDER BY created_at DESC
	`

	var users []*domain.User
	err := r.db.Select(&users, query)
	if err != nil {
		logger.Error("Failed to get active users", logger.ErrorField(err))
		return nil, fmt.Errorf("failed to get active users: %w", err)
	}

	return users, nil
}

// GetUsersByLevel retrieves users by level
func (r *userRepository) GetUsersByLevel(level int) ([]*domain.User, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, phone,
			upline_id, level, is_active, is_verified, balance, credit_limit,
			markup_percentage, allow_debt, max_daily_transaction,
			created_at, updated_at, last_login_at
		FROM users WHERE level = $1 ORDER BY created_at DESC
	`

	var users []*domain.User
	err := r.db.Select(&users, query, level)
	if err != nil {
		logger.Error("Failed to get users by level", 
			logger.Int("level", level),
			logger.ErrorField(err),
		)
		return nil, fmt.Errorf("failed to get users by level: %w", err)
	}

	return users, nil
}
