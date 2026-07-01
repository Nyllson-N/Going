// Package http é a camada de transporte (delivery) HTTP da aplicação.
package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"e/internal/domain"
	"e/internal/httputil"
	"e/internal/usecase"
)

// UserHandler expõe os endpoints HTTP do domínio de usuários.
type UserHandler struct {
	uc *usecase.UserUseCase
}

// NewUserHandler cria o handler de usuários.
func NewUserHandler(uc *usecase.UserUseCase) *UserHandler {
	return &UserHandler{uc: uc}
}

// RegisterRoutes registra as rotas de usuários no mux (roteamento por método, Go 1.22+).
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	mux.HandleFunc("GET /users", h.List)
}

// DTOs de entrada/saída em JSON — concern exclusivo da camada de transporte.
type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func toResponse(u domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// Register cria um novo usuário.
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	u, err := h.uc.Register(r.Context(), usecase.RegisterInput(req))
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.JSON(w, http.StatusCreated, toResponse(*u))
}

// Login autentica um usuário pelas credenciais.
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, http.StatusBadRequest, "JSON inválido")
		return
	}
	u, err := h.uc.Login(r.Context(), usecase.LoginInput(req))
	if err != nil {
		writeError(w, err)
		return
	}
	httputil.JSON(w, http.StatusOK, toResponse(*u))
}

// List retorna todos os usuários.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.uc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	res := make([]userResponse, len(users))
	for i, u := range users {
		res[i] = toResponse(u)
	}
	httputil.JSON(w, http.StatusOK, res)
}

// writeError mapeia erros de domínio para os status HTTP adequados.
func writeError(w http.ResponseWriter, err error) {
	var validation domain.ValidationError
	switch {
	case errors.As(err, &validation):
		httputil.Error(w, http.StatusBadRequest, validation.Message)
	case errors.Is(err, domain.ErrEmailTaken):
		httputil.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		httputil.Error(w, http.StatusUnauthorized, err.Error())
	default:
		httputil.Error(w, http.StatusInternalServerError, "erro interno do servidor")
	}
}
