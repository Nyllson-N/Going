// Package usecase contém as regras de negócio (casos de uso) da aplicação.
// Depende apenas das interfaces definidas em domain, nunca de implementações.
package usecase

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"e/internal/domain"
)

const (
	// bcryptCost define o custo do hash. 12 equilibra segurança e desempenho.
	bcryptCost = 12
	// minPasswordLen é o tamanho mínimo aceito para a senha.
	minPasswordLen = 8
)

// RegisterInput são os dados de entrada para registrar um usuário.
type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

// LoginInput são os dados de entrada para autenticar um usuário.
type LoginInput struct {
	Email    string
	Password string
}

// UserUseCase concentra as regras de negócio relacionadas a usuários.
type UserUseCase struct {
	repo domain.UserRepository
}

// NewUserUseCase cria o caso de uso de usuários, injetando o repositório.
func NewUserUseCase(repo domain.UserRepository) *UserUseCase {
	return &UserUseCase{repo: repo}
}

// Register valida os dados, gera o hash bcrypt da senha e persiste o usuário.
func (uc *UserUseCase) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	name := strings.TrimSpace(in.Name)
	email := normalizeEmail(in.Email)

	if name == "" || email == "" {
		return nil, domain.ValidationError{Message: "nome e e-mail são obrigatórios"}
	}
	if len(in.Password) < minPasswordLen {
		return nil, domain.ValidationError{Message: "a senha deve ter no mínimo 8 caracteres"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcryptCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         "user",
	}
	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login valida as credenciais comparando a senha com o hash bcrypt armazenado.
// Sempre retorna ErrInvalidCredentials para e-mail inexistente ou senha errada,
// evitando revelar se um e-mail está ou não cadastrado.
func (uc *UserUseCase) Login(ctx context.Context, in LoginInput) (*domain.User, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	u, err := uc.repo.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(in.Password)) != nil {
		return nil, domain.ErrInvalidCredentials
	}
	return u, nil
}

// List retorna todos os usuários.
func (uc *UserUseCase) List(ctx context.Context) ([]domain.User, error) {
	return uc.repo.List(ctx)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
