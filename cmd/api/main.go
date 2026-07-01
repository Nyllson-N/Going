package main

import (
	"log"
	"net/http"
	"time"

	"e/internal/config"
	"e/internal/database"
	transporthttp "e/internal/transport/http"
)

func main() {
	cfg := config.New()

	// Conexão com o banco de dados.
	db, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer db.Close()
	log.Println("Banco de dados conectado")

	// Migrações (idempotentes).
	if err := database.Migrate(db); err != nil {
		log.Fatalf("erro ao migrar banco: %v", err)
	}
	log.Println("Migração do banco concluída")

	// Monta o handler HTTP (rotas + middlewares + injeção de dependências).
	handler := transporthttp.NewRouter(cfg, db)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Servidor ouvindo em %s", srv.Addr)
	log.Fatal(srv.ListenAndServe())
}
