// Package postgres implementa os repositórios de domain sobre PostgreSQL.
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"e/internal/domain"
)

// UserStore é a implementação (adapter) de domain.UserRepository sobre PostgreSQL.
type UserStore struct {
	db *sql.DB
}

// Garantia em tempo de compilação de que UserStore satisfaz o contrato.
var _ domain.UserRepository = (*UserStore)(nil)

// NewUserStore cria o store de usuários.
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, u *domain.User) error {
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO users (name, email, password_hash, role)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at`,
		u.Name, u.Email, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt)
	if isUniqueViolation(err) {
		return domain.ErrEmailTaken
	}
	return err
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, email, password_hash, role, created_at
		 FROM users WHERE email = $1`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *UserStore) List(ctx context.Context) ([]domain.User, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, email, password_hash, role, created_at
		 FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// isUniqueViolation detecta a violação de unicidade do PostgreSQL (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
