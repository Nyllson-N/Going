// Package database cuida da conexão e das migrações do banco de dados.
package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// Connect abre a conexão com o PostgreSQL (driver lib/pq), configura o pool
// de conexões e valida a conexão com um Ping.
func Connect(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("database: erro ao abrir conexão: %w", err)
	}

	// Configuração do pool de conexões.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database: erro ao conectar: %w", err)
	}

	return db, nil
}

// Migrate garante que a tabela de usuários exista com as colunas necessárias.
// É idempotente: cria a tabela se não existir e adiciona as colunas
// password_hash (hash bcrypt) e role caso a tabela já exista sem elas.
func Migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          TEXT        NOT NULL,
    email         TEXT        NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '';`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user';`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("database: erro ao migrar schema: %w", err)
		}
	}
	return nil
}
