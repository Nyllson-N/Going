// Package domain contém as entidades de negócio e seus contratos (interfaces).
// Não depende de nenhuma camada externa (banco, HTTP, etc.).
package domain

import (
	"context"
	"errors"
	"time"
)

// User é a entidade de domínio que representa um usuário.
// PasswordHash nunca deve ser exposto na camada de transporte.
type User struct {
	ID           int64
	Name         string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

// UserRepository é o contrato (porta) de persistência de usuários.
// As implementações (adapters) ficam em internal/store.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context) ([]User, error)
}

// Erros de domínio, reaproveitados pelas demais camadas.
var (
	// ErrEmailTaken indica e-mail já cadastrado (HTTP 409).
	ErrEmailTaken = errors.New("e-mail já cadastrado")
	// ErrInvalidCredentials indica e-mail ou senha inválidos no login (HTTP 401).
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	// ErrNotFound indica que o usuário não foi encontrado.
	ErrNotFound = errors.New("usuário não encontrado")
)

// ValidationError representa um erro de validação de entrada (HTTP 400).
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }
