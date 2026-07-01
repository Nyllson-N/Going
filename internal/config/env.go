package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Load lê um arquivo .env e define as variáveis no ambiente do processo.
// Por padrão não sobrescreve variáveis já existentes no ambiente.
// Se nenhum caminho for informado, usa ".env".
func Load(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}
	for _, path := range paths {
		if err := loadFile(path, false); err != nil {
			return err
		}
	}
	return nil
}

// Overload funciona como Load, mas sobrescreve variáveis já existentes.
func Overload(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}
	for _, path := range paths {
		if err := loadFile(path, true); err != nil {
			return err
		}
	}
	return nil
}

func loadFile(path string, override bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("config: não foi possível abrir %q: %w", path, err)
	}
	defer file.Close()

	values, err := parse(file)
	if err != nil {
		return fmt.Errorf("config: erro ao ler %q: %w", path, err)
	}

	for key, value := range values {
		if !override {
			if _, exists := os.LookupEnv(key); exists {
				continue
			}
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("config: erro ao definir %q: %w", key, err)
		}
	}
	return nil
}

// parse lê linhas no formato KEY=VALUE, ignorando comentários e linhas vazias.
func parse(file *os.File) (map[string]string, error) {
	values := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Ignora linhas vazias e comentários.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Suporta o prefixo "export KEY=VALUE".
		line = strings.TrimPrefix(line, "export ")

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}

		values[key] = cleanValue(strings.TrimSpace(value))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

// cleanValue remove aspas ao redor do valor e comentários inline em valores não citados.
func cleanValue(value string) string {
	if len(value) >= 2 {
		first, last := value[0], value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			return value[1 : len(value)-1]
		}
	}

	// Remove comentário inline em valores sem aspas (ex: PORT=8080 # comentário).
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}
	return value
}
