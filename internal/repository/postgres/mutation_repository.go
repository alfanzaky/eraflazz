package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/alfanzaky/eraflazz/internal/domain"
	"github.com/alfanzaky/eraflazz/pkg/logger"
)

type mutationRepository struct {
	db *sqlx.DB
}

// NewMutationRepository creates a new mutation repository instance
func NewMutationRepository(db *sqlx.DB) domain.MutationRepository {
	return &mutationRepository{db: db}
}

func (r *mutationRepository) Create(mutation *domain.Mutation) error {
	query := `
        INSERT INTO mutations (
            id, user_id, type, amount, balance_before, balance_after,
            reference_type, reference_id, description, notes,
            created_by, ip_address, user_agent, created_at
        ) VALUES (
            :id, :user_id, :type, :amount, :balance_before, :balance_after,
            :reference_type, :reference_id, :description, :notes,
            :created_by, :ip_address, :user_agent, NOW()
        )`

	_, err := r.db.NamedExec(query, mutation)
	if err != nil {
		logger.Error("Failed to create mutation", logger.ErrorField(err))
		return fmt.Errorf("failed to create mutation: %w", err)
	}

	return nil
}

func (r *mutationRepository) GetByID(id string) (*domain.Mutation, error) {
	query := `SELECT * FROM mutations WHERE id = $1`
	var mutation domain.Mutation
	err := r.db.Get(&mutation, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("mutation not found")
		}
		return nil, fmt.Errorf("failed to get mutation: %w", err)
	}
	return &mutation, nil
}

func (r *mutationRepository) GetByUserID(userID string, limit, offset int) ([]*domain.Mutation, error) {
	query := `
        SELECT * FROM mutations
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3`

	var mutations []*domain.Mutation
	err := r.db.Select(&mutations, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user mutations: %w", err)
	}
	return mutations, nil
}

func (r *mutationRepository) GetByReference(referenceType, referenceID string) ([]*domain.Mutation, error) {
	query := `
        SELECT * FROM mutations
        WHERE reference_type = $1 AND reference_id = $2
        ORDER BY created_at DESC`

	var mutations []*domain.Mutation
	err := r.db.Select(&mutations, query, referenceType, referenceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutations by reference: %w", err)
	}
	return mutations, nil
}

func (r *mutationRepository) GetBalanceHistory(userID string, limit, offset int) ([]*domain.Mutation, error) {
	return r.GetByUserID(userID, limit, offset)
}

func (r *mutationRepository) GetCurrentBalance(userID string) (float64, error) {
	query := `
        SELECT balance_after
        FROM mutations
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 1`

	var balance float64
	err := r.db.Get(&balance, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get current balance: %w", err)
	}
	return balance, nil
}
