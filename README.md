# gostart

CLI para scaffolding de projetos Go REST API com arquitetura modular, **Security-First** e **API-First**.

```
gostart init minha-api --module github.com/acme/minha-api
```

## Instalação

### Via `go install`

```bash
go install github.com/marcelofabianov/gostart@latest
```

### Build manual

```bash
git clone https://github.com/marcelofabianov/gostart.git
cd gostart
go build -o gostart .
sudo mv gostart /usr/local/bin/
```

## Uso

### Modo interativo (sem flags)

```bash
gostart init
```

Abre prompts para preencher nome do projeto, módulo Go, banco de dados, cache, Docker e CI.

### Modo direto (com flags)

```bash
gostart init <project-name> --module <module-path> [flags]
```

**Exemplos:**

```bash
# Projeto completo com todas as features
gostart init payments --module github.com/acme/payments

# Sem Docker (para quem usa Kubernetes/Helm)
gostart init user-service --module github.com/acme/user-service --no-docker

# Sem CI (GitHub Actions)
gostart init orders --module github.com/acme/orders --no-ci

# Sem banco de dados nem cache
gostart init notifier --module github.com/acme/notifier --db=none --no-cache
```

## Flags

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--module` | — | Caminho do módulo Go (ex: `github.com/user/projeto`) |
| `--db` | `postgres` | Banco de dados: `postgres` ou `none` |
| `--no-cache` | `false` | Omite Redis/cache (`pkg/cache`) |
| `--no-docker` | `false` | Omite `Dockerfile` e `docker-compose.yml` |
| `--no-ci` | `false` | Omite `.github/workflows/ci.yml` |

## Estrutura gerada

```
minha-api/
├── cmd/api/
│   └── main.go                    # Entrypoint — fx.New()
├── config/
│   └── config.go                  # Viper + structs de configuração
├── internal/
│   ├── di/
│   │   ├── pkg.go                 # PkgModule (Config, Logger, DB, Cache, Crypto)
│   │   ├── app.go                 # AppModule (Router Chi, Server TLS)
│   │   └── hello.go               # HelloModule
│   └── hello/                     # Bounded context de exemplo
│       ├── domain/entity.go
│       ├── handler/handler.go     # Implementa web.Router → GET /api/v1/hello
│       ├── port/port.go           # Interfaces
│       ├── usecase/usecase.go
│       ├── storage/               # Implementações de repositório
│       └── publisher/             # Publicadores de eventos
├── pkg/
│   ├── web/
│   │   ├── server.go              # HTTPS TLS 1.2/1.3
│   │   ├── health.go              # GET /health + /health/ready
│   │   ├── router.go              # Interface web.Router
│   │   ├── response.go            # Helpers de resposta JSON
│   │   ├── context.go
│   │   ├── chi/router.go          # Chi + middleware stack completo
│   │   └── middleware/            # 14 middlewares de segurança
│   ├── database/postgres.go       # pgx v5 + pool + retry
│   ├── cache/cache.go             # go-redis v9 + pool
│   ├── logger/logger.go           # slog estruturado
│   ├── retry/                     # Retry com backoff exponencial
│   ├── token/token.go             # JWT access + refresh
│   ├── crypto/hasher.go           # Argon2id
│   └── validation/                # go-playground/validator
├── db/migrations/                 # SQL migrations (goose)
├── docs/docs.go                   # Stub para swag (OpenAPI)
├── scripts/gen-certs.sh           # Gera certs TLS self-signed para dev
├── docker/Dockerfile              # Multi-stage distroless
├── docker-compose.yml             # Postgres + Redis + App
├── Makefile                       # Targets de build, test, lint, docs
├── .github/workflows/ci.yml       # golangci-lint + gosec + test + build
├── .golangci.yml
├── .env.example
└── .gitignore
```

## Stack do projeto gerado

| Camada | Biblioteca |
|--------|-----------|
| DI | `go.uber.org/fx` |
| Roteamento | `github.com/go-chi/chi/v5` |
| Configuração | `github.com/spf13/viper` |
| Banco | `github.com/jackc/pgx/v5` |
| Cache | `github.com/redis/go-redis/v9` |
| JWT | `github.com/golang-jwt/jwt/v5` |
| Cripto | `golang.org/x/crypto` (Argon2id) |
| Validação | `github.com/go-playground/validator/v10` |
| Migrations | `github.com/pressly/goose/v3` |
| Erros | `github.com/marcelofabianov/fault` |
| Rate limit | `github.com/go-redis/redis_rate/v10` + `sony/gobreaker` |
| Logging | `log/slog` (stdlib) |

## Middleware stack (ordem de execução)

```
Recovery → RequestID → RealIP → Logger → SecurityHeaders → HTTPSOnly
  → CORS → RequestSize → Compression → RateLimit(CircuitBreaker)
  → /ping /health /health/ready
  → /api/v1/
    → Timeout → AcceptJSON → AllowContentType → CSRF → [rotas]
```

## Primeiros passos após geração

```bash
cd minha-api

# 1. Configurar variáveis de ambiente
cp .env.example .env

# 2. Gerar certificados TLS para desenvolvimento
make certs

# 3. Subir infraestrutura e aplicação
make run

# 4. Verificar saúde da API
curl -k https://localhost:8443/health

# 5. Testar endpoint de exemplo
curl -k https://localhost:8443/api/v1/hello
```

## Adicionar novo bounded context

Crie o módulo seguindo a estrutura do `hello`:

```
internal/payments/
├── domain/         # Entidades e value objects
├── handler/        # Implementa web.Router (RegisterRoutes)
├── port/           # Interfaces (UseCases, Repositories)
├── usecase/        # Lógica de negócio
├── storage/        # Implementações de repositório
└── publisher/      # Publicadores de eventos
```

Registre no DI em `internal/di/`:

```go
// internal/di/payments.go
var PaymentsModule = fx.Module("payments",
    fx.Provide(
        usecase.NewProcessPaymentUseCase,
        AsRouter(handler.NewPaymentsHandler),
    ),
)
```

Adicione `di.PaymentsModule` em `cmd/api/main.go`.

## Makefile targets

| Target | Descrição |
|--------|-----------|
| `make run` | Sobe docker-compose (infra + app) |
| `make build` | Compila o binário |
| `make test` | Executa testes com race detector |
| `make lint` | golangci-lint |
| `make sec` | gosec (análise de segurança) |
| `make docs` | Gera documentação OpenAPI (swag) |
| `make certs` | Gera certificados TLS self-signed |
| `make migrate-up` | Aplica migrations |
| `make migrate-down` | Reverte última migration |
| `make tidy` | go mod tidy |

## Contribuição

```bash
git clone https://github.com/marcelofabianov/gostart.git
cd gostart
go mod download
go test ./...
go build .
```

Templates ficam em `internal/generator/templates/`. Use `[[` e `]]` como delimitadores em vez de `{{` `}}` para evitar conflito com código Go nos templates.

## Licença

MIT
