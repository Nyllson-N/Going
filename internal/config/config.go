package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config reúne todas as configurações da aplicação, lidas do ambiente.
type Config struct {
	Port       string
	CORSOrigin string
	RateLimit  int
	RatePeriod time.Duration
	DB         DBConfig
}

// DBConfig guarda os dados de conexão com o PostgreSQL.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN monta a string de conexão no formato do PostgreSQL.
func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// New carrega o arquivo .env (se existir) e monta a Config a partir do ambiente,
// aplicando valores padrão quando uma variável não está definida.
func New() Config {
	// Carrega .env sem sobrescrever variáveis já presentes no ambiente.
	_ = Load()

	return Config{
		Port:       getEnv("PORT", "8080"),
		CORSOrigin: getEnv("CORS_ORIGIN", "*"),
		RateLimit:  getEnvInt("RATE_LIMIT", 100),
		RatePeriod: getEnvDuration("RATE_PERIOD", time.Second),
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "golangbackend"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return def
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	if v, err := time.ParseDuration(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return def
}
