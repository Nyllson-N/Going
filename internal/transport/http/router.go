package http

import (
	"database/sql"
	"net/http"

	"e/internal/config"
	"e/internal/middleware"
	"e/internal/store/postgres"
	"e/internal/usecase"
)

// NewRouter monta o handler HTTP completo da aplicação e faz a injeção de
// dependências de cada domínio: store (adapter) -> usecase -> handler.
// Em seguida aplica os middlewares globais.
func NewRouter(cfg config.Config, db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	// Health check.
	mux.HandleFunc("GET /health", healthHandler(db))

	// Domínio: usuários.
	userStore := postgres.NewUserStore(db)
	userUseCase := usecase.NewUserUseCase(userStore)
	NewUserHandler(userUseCase).RegisterRoutes(mux)

	// Middlewares aplicados de fora para dentro: CORS -> rate limit -> rotas.
	limiter := middleware.NewRateLimiter(cfg.RateLimit, cfg.RatePeriod)
	return middleware.CORS(cfg.CORSOrigin, limiter.Handler(mux))
}

// healthHandler verifica a saúde da aplicação, incluindo a conexão com o banco.
func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "banco indisponível", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}
}
