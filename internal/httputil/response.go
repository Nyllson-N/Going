// Package httputil reúne helpers para respostas HTTP em JSON.
package httputil

import (
	"encoding/json"
	"net/http"
)

// JSON escreve uma resposta JSON (indentada com 2 espaços) com o status informado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(v)
	}
}

// Error escreve uma resposta de erro padronizada: {"error": "mensagem"}.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
