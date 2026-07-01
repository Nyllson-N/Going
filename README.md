# GolangBackend

API HTTP em Go para cadastro e autenticação de usuários, organizada em
**Clean Architecture (ports & adapters)**. As senhas são armazenadas com **bcrypt**
e os dados ficam em **PostgreSQL**.

---

## Sumário

- [Tecnologias](#tecnologias)
- [Estrutura do projeto](#estrutura-do-projeto)
- [A regra de dependência](#a-regra-de-dependência-por-que-isso-é-robusto)
- [Como rodar](#como-rodar)
- [Configuração (.env)](#configuração-env)
- [Endpoints da API](#endpoints-da-api)
- [Segurança](#segurança)
- [Como estender](#como-estender)

---

## Tecnologias

- **Go 1.25+** — usa o roteamento por método do `net/http` (Go 1.22+), sem framework externo.
- **PostgreSQL** — driver [`github.com/lib/pq`](https://github.com/lib/pq).
- **bcrypt** — [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) para hash de senhas.
- **Docker Compose** — sobe o banco para desenvolvimento.

---

## Estrutura do projeto

```
cmd/api/main.go                       → entrypoint: só "liga" as peças
internal/
  domain/
    user.go        → entidade User + interface UserRepository + erros de domínio
  usecase/
    user.go        → UserUseCase: regras de negócio + bcrypt (depende só de domain)
  store/
    postgres/
      user.go      → UserStore: adapter que IMPLEMENTA domain.UserRepository
  transport/
    http/
      user.go      → UserHandler: DTOs JSON + tradução de erro → status HTTP
      router.go    → monta rotas, middlewares e injeta as dependências
  config/          → Config + leitor de .env
  database/        → Connect + Migrate
  middleware/      → CORS + rate limit
  httputil/        → helpers de resposta JSON (indentado com 2 espaços)
```

Cada camada tem **uma responsabilidade**:

| Camada            | Responsabilidade                                                        |
| ----------------- | ----------------------------------------------------------------------- |
| `domain`          | Núcleo do negócio: a entidade `User`, o contrato `UserRepository` e os erros. Não importa ninguém. |
| `usecase`         | Regras de negócio (validação, hash da senha, orquestração). Depende só das *interfaces* de `domain`. |
| `store/postgres`  | Adapter de persistência: implementa `domain.UserRepository` em SQL.     |
| `transport/http`  | Camada de entrega: converte JSON ↔ tipos do usecase e mapeia erros para status HTTP. |
| `config`          | Carrega configuração do `.env`/ambiente com valores padrão.             |
| `database`        | Abre a conexão e roda as migrações (idempotentes).                      |
| `middleware`      | CORS e rate limit por IP.                                               |
| `httputil`        | Helpers `JSON()` e `Error()` para respostas padronizadas.              |

---

## A regra de dependência (por que isso é "robusto")

O fluxo de dependências aponta sempre **para dentro**, em direção ao domínio:

```
transport/http  ─→  usecase  ─→  domain  ←─  store/postgres
   (HTTP/JSON)      (negócio)   (núcleo)      (PostgreSQL)
```

- **`domain`** não importa nenhuma outra camada. Define **o que** é um `User` e o
  contrato `UserRepository` (a *porta*).
- **`usecase`** depende apenas da *interface* `domain.UserRepository` — ele não sabe
  que existe PostgreSQL. Por isso é **testável com um mock** do repositório.
- **`store/postgres`** *implementa* a interface (tem
  `var _ domain.UserRepository = (*UserStore)(nil)`, que garante isso em tempo de
  compilação).
- **`transport/http`** converte JSON ↔ tipos do usecase e traduz erros de domínio em
  status HTTP. A entidade `domain.User` **nunca vaza a senha**: a resposta usa um DTO
  `userResponse` separado.

A "amarração" (qual store concreto entra em qual usecase) acontece em **um único
lugar**: `internal/transport/http/router.go`.

```go
userStore   := postgres.NewUserStore(db)        // adapter concreto
userUseCase := usecase.NewUserUseCase(userStore) // recebe a interface
NewUserHandler(userUseCase).RegisterRoutes(mux)  // expõe via HTTP
```

Trocar o PostgreSQL por outro banco = criar um novo `store/` e mudar essa linha.

---

## Como rodar

### 1. Subir o banco (Docker)

```bash
docker compose up -d
```

### 2. Rodar a API

```bash
go run ./cmd/api
```

A aplicação roda as migrações automaticamente no startup (cria/ajusta a tabela
`users`) e sobe em `http://localhost:8080`.

> **Build:**
> ```bash
> go build -o bin/api ./cmd/api && ./bin/api
> ```

> **IDE (GoLand/VS Code):** o entrypoint é `cmd/api`. Configure a Run Configuration
> para o pacote `e/cmd/api` (ou diretório `cmd/api`).

---

## Configuração (.env)

As variáveis são lidas do ambiente; o arquivo `.env` é carregado automaticamente.
Valores padrão são aplicados quando uma variável não está definida.

| Variável      | Padrão           | Descrição                                              |
| ------------- | ---------------- | ------------------------------------------------------ |
| `PORT`        | `8080`           | Porta HTTP da API.                                     |
| `CORS_ORIGIN` | `*`              | Origem permitida no CORS.                              |
| `RATE_LIMIT`  | `100`            | Número de requisições por janela de tempo.            |
| `RATE_PERIOD` | `1s`             | Tamanho da janela (`ms`, `s`, `m`, `h`).              |
| `DB_HOST`     | `localhost`      | Host do PostgreSQL.                                    |
| `DB_PORT`     | `5432`           | Porta do PostgreSQL.                                   |
| `DB_USER`     | `postgres`       | Usuário do banco.                                      |
| `DB_PASSWORD` | `postgres`       | Senha do banco.                                        |
| `DB_NAME`     | `golangbackend`  | Nome do banco.                                         |
| `DB_SSLMODE`  | `disable`        | Modo SSL da conexão (`disable`, `require`, etc.).     |

---

## Endpoints da API

Todas as respostas são em JSON **indentado com 2 espaços**. Respostas de erro seguem
o formato `{ "error": "mensagem" }`.

### `GET /health`

Verifica a saúde da aplicação (inclui `Ping` no banco).

```bash
curl localhost:8080/health
# ok
```

### `POST /register`

Cria um novo usuário (senha mínima de 8 caracteres).

```bash
curl -X POST localhost:8080/register \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ana","email":"ana@x.com","password":"senha12345"}'
```

**201 Created**
```json
{
  "id": 1,
  "name": "Ana",
  "email": "ana@x.com",
  "role": "user",
  "created_at": "2026-06-30T22:54:31.922486Z"
}
```

| Situação                       | Status |
| ------------------------------ | ------ |
| Criado com sucesso             | `201`  |
| JSON inválido / campos vazios  | `400`  |
| Senha com menos de 8 chars     | `400`  |
| E-mail já cadastrado           | `409`  |

### `POST /login`

Valida as credenciais comparando a senha com o hash bcrypt.

```bash
curl -X POST localhost:8080/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"ana@x.com","password":"senha12345"}'
```

| Situação                          | Status |
| --------------------------------- | ------ |
| Autenticado                       | `200`  |
| Credenciais inválidas             | `401`  |

> Por segurança, e-mail inexistente e senha errada retornam **a mesma** mensagem
> (`credenciais inválidas`), para não revelar se um e-mail está cadastrado.

### `GET /users`

Lista todos os usuários (sem expor senha/hash).

```bash
curl localhost:8080/users
```

> Requisições com o método errado (ex.: `GET /register`) retornam **405 Method Not
> Allowed** automaticamente, graças ao roteamento por método do `net/http`.

---

## Segurança

- **Senhas com bcrypt** (custo 12, com *salt* embutido). O hash fica na coluna
  `password_hash`; a senha em texto puro nunca é armazenada nem retornada.
- **DTO de resposta separado da entidade** — o `password_hash` jamais é serializado
  para o cliente.
- **Mensagens de erro genéricas no login** para evitar enumeração de e-mails.
- **Rate limit por IP** e **CORS** configuráveis via `.env`.
- **Timeouts no servidor HTTP** (Read/Write/Idle) para resiliência.

---

## Como estender

Para adicionar um novo domínio (ex.: `product`), siga o mesmo fluxo:

1. **`domain/product.go`** — entidade `Product` + interface `ProductRepository` + erros.
2. **`usecase/product.go`** — regras de negócio dependendo da interface.
3. **`store/postgres/product.go`** — implementação SQL do repositório.
4. **`transport/http/product.go`** — handler + DTOs JSON.
5. **`transport/http/router.go`** — registrar as rotas e injetar as dependências.

Como o `usecase` depende de uma *interface*, dá para testá-lo com um mock do
repositório, sem subir banco.
