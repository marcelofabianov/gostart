# `gostart` — Go REST API Scaffolding CLI

> CLI que gera projetos Go para API REST de forma estruturada, segura e com alto padrão de qualidade.
> Com um simples comando `gostart init`, toda a estrutura do projeto é criada pronta para produção.

---

## Objetivo

Eliminar o boilerplate repetitivo na criação de novas APIs Go, gerando uma estrutura **completa, segura e consistente** baseada em padrões comprovados (referência: `personal-todo`).

---

## Princípios

| Princípio         | Descrição                                                                 |
|-------------------|---------------------------------------------------------------------------|
| **Security-First** | TLS 1.2/1.3, HTTPS enforcement, SecurityHeaders, CSRF, RateLimit, Argon2 |
| **API-First**      | `/health`, `/health/ready`, `/api/v1/`, Swagger via swaggo               |
| **Modular**        | Bounded Contexts em `internal/{modulo}/` com DI via `go.uber.org/fx`     |
| **Zero boilerplate** | `gostart init` → projeto compilável e funcional imediatamente          |

---

## Uso

```bash
# instalação
go install github.com/marcelofabianov/gostart@latest

# criar novo projeto (interativo se omitir flags)
gostart init minha-api \
  --module github.com/user/minha-api \
  --db postgres \
  --no-cache=false

# flags disponíveis
gostart init --help
```

### Flags do `init`

| Flag          | Padrão        | Descrição                              |
|---------------|---------------|----------------------------------------|
| `--module`    | (obrigatório) | Caminho do módulo Go                   |
| `--db`        | `postgres`    | `postgres` ou `none`                   |
| `--no-cache`  | `false`       | Omite Redis / pkg/cache                |
| `--no-docker` | `false`       | Omite Dockerfile + docker-compose      |
| `--no-ci`     | `false`       | Omite GitHub Actions workflow          |

---

## Estrutura do CLI (`gostart`)

```
framework/
├── cmd/
│   ├── root.go               # cobra root, versão, flags globais
│   └── init.go               # comando `init` com todas as flags
├── internal/
│   └── generator/
│       ├── generator.go      # orquestra criação de arquivos via embed.FS + text/template
│       ├── options.go        # struct ProjectOptions com todas as configurações
│       └── embed.go          # //go:embed templates/** — embute no binário
├── templates/                # todos os arquivos .tmpl do projeto gerado
├── _docs/                    # documentação do CLI
│   └── PLANNING.md
├── main.go
├── go.mod
└── README.md
```

---

## Projeto Gerado — Estrutura Completa

```
{project}/
│
├── cmd/
│   └── api/
│       └── main.go               # fx.New() → PkgModule + AppModule + HelloModule
│
├── config/
│   ├── config.go                 # structs: Config, HTTPConfig, TLSConfig, DBConfig...
│   └── defaults.go               # setDefaults() via Viper
│
├── internal/
│   ├── di/
│   │   ├── pkg.go                # PkgModule: Config, Logger, DB, Cache, Validation
│   │   ├── app.go                # AppModule: Router, Server, AsRouter(), AsHealthChecker()
│   │   └── hello.go              # HelloModule: exemplo de módulo wired
│   │
│   └── hello/                    # ← Bounded Context de exemplo (Hello World)
│       ├── domain/
│       │   └── entity.go         # entidade Hello
│       ├── handler/
│       │   └── handler.go        # GET /api/v1/hello → implementa web.Router
│       ├── port/
│       │   └── port.go           # interface SayHelloPort
│       ├── usecase/
│       │   └── usecase.go        # SayHelloUseCase
│       ├── storage/
│       │   └── .gitkeep          # placeholder para implementações de repositório
│       └── publisher/
│           └── .gitkeep          # placeholder para publishers de eventos
│
├── pkg/
│   ├── logger/
│   │   ├── logger.go             # wrapper log/slog: Debug, Info, Warn, Error, With
│   │   └── config.go             # LogLevel, LogFormat, Config
│   │
│   ├── web/
│   │   ├── server.go             # HTTP Server + TLS 1.2/1.3 + graceful shutdown
│   │   ├── health.go             # /health (liveness) + /health/ready (readiness)
│   │   ├── response.go           # helpers: Success, Error, Created, NoContent...
│   │   ├── router.go             # interface Router { RegisterRoutes(chi.Router) }
│   │   ├── context.go            # GetLogger, GetRequestID, GetUserID helpers
│   │   └── chi/
│   │       └── router.go         # NewRouter() — stack completa de middlewares
│   │
│   ├── web/middleware/
│   │   ├── security_headers.go   # CSP, HSTS, X-Frame-Options, Referrer-Policy...
│   │   ├── https_only.go         # enforça HTTPS, rejeita HTTP
│   │   ├── csrf.go               # HMAC-SHA256 double-submit cookie
│   │   ├── rate_limit.go         # Redis rate limit + Sony Circuit Breaker
│   │   ├── security_logger.go    # log estruturado de eventos de segurança
│   │   ├── logger.go             # request logging com latência e status
│   │   ├── recovery.go           # panic recovery com stack trace
│   │   ├── request_id.go         # X-Request-ID header
│   │   ├── real_ip.go            # extração de IP real via trusted proxies
│   │   ├── cors.go               # CORS configurável por origem/método
│   │   ├── timeout.go            # request timeout configurável
│   │   ├── accept.go             # Accept: application/json enforcement
│   │   ├── sanitize.go           # sanitização de input (bluemonday)
│   │   └── request_size.go       # limite de tamanho do body
│   │
│   ├── database/
│   │   └── postgres.go           # pgx v5 + connection pool + retry + health check
│   │
│   ├── cache/
│   │   └── cache.go              # go-redis v9 + pool + health check (opcional)
│   │
│   ├── token/
│   │   └── token.go              # JWT access + refresh (golang-jwt/jwt v5)
│   │
│   ├── crypto/
│   │   └── hasher.go             # Argon2id hasher (hash + verify)
│   │
│   ├── validation/
│   │   ├── validator.go          # go-playground/validator + fault + campo sensível redact
│   │   └── config.go             # ValidationConfig (logging, sensitive fields)
│   │
│   └── retry/
│       ├── retry.go              # retry com contexto e backoff
│       └── backoff_strategy.go   # exponential backoff com jitter
│
├── db/
│   └── migrations/
│       ├── .gitkeep
│       └── 00001_init.sql        # migration inicial (esqueleto)
│
├── scripts/
│   └── gen-certs.sh              # gera certs/server.crt + certs/server.key (dev TLS)
│
├── docker/
│   └── Dockerfile                # multi-stage: builder (go) + runtime (distroless/scratch)
│
├── .github/
│   └── workflows/
│       └── ci.yml                # lint (golangci-lint) + gosec + test -race + build
│
├── docker-compose.yml            # postgres + redis + app (com healthchecks)
├── Makefile                      # todos os targets de desenvolvimento
├── .golangci.yml                 # config do linter
├── .env.example                  # todos APP_* vars documentados com valores padrão
├── .gitignore
├── go.mod
└── README.md
```

---

## Arquitetura Modular — Bounded Contexts

Cada módulo de negócio segue a estrutura do `personal-todo`:

```
internal/{modulo}/
├── domain/         # entidades e value objects do domínio
├── handler/        # handlers HTTP — implementam web.Router
├── port/           # interfaces (ports): UseCases, Repositories
├── usecase/        # regras de negócio (application layer)
├── storage/        # implementações de repositório (DB, cache)
└── publisher/      # publicação de eventos (NATS, etc.)
```

### Adicionando um novo módulo

```go
// 1. Criar internal/user/ com a estrutura acima
// 2. Criar internal/di/user.go
var UserModule = fx.Module("user",
    fx.Provide(
        storage.NewPostgresUserRepository,
        func(r *storage.PostgresUserRepository) port.UserRepositoryPort { return r },
        fx.Annotate(usecase.NewCreateUserUseCase, fx.As(new(port.CreateUserUseCase))),
        AsRouter(handler.NewUserHandler),
    ),
)

// 3. Adicionar no main.go
fx.New(
    di.PkgModule,
    di.AppModule,
    di.HelloModule,  // exemplo
    di.UserModule,   // novo módulo
    fx.Invoke(func(*web.Server) {}),
).Run()
```

---

## DI com go.uber.org/fx

```
main.go
  └── fx.New()
        ├── PkgModule       → Config, Logger, DB, Cache, Validator, Crypto...
        ├── AppModule       → Chi Router, Web Server (TLS), HealthCheckers
        └── HelloModule     → SayHelloUseCase, HelloHandler (registrado como Router)
```

O `AppModule` coleta automaticamente todos os `web.Router` via `group:"routers"`:

```go
// qualquer módulo pode registrar rotas assim:
AsRouter(handler.NewHelloHandler)  // → injetado no chi Router automaticamente
```

---

## Rotas geradas

```
GET  /ping             → heartbeat (chi middleware)
GET  /health           → liveness  {"status":"healthy","uptime":"..."}
GET  /health/ready     → readiness {"status":"healthy","checks":{"database":...}}
GET  /swagger/*        → documentação OpenAPI

GET  /api/v1/hello     → Hello World (módulo de exemplo)
```

---

## Middleware Stack (Security-First)

```
[Global]
  Recovery            ← panic recovery com log
  RequestID           ← X-Request-ID
  RealIP              ← IP real via trusted proxies
  Logger              ← request/response log estruturado
  SecurityHeaders     ← CSP, HSTS, X-Frame-Options, Referrer-Policy...
  HTTPSOnly           ← rejeita HTTP se TLS habilitado
  CORS                ← origens/métodos configuráveis
  RequestSize         ← limite de body (default: 1MB)
  Compression         ← gzip configurável
  RateLimit           ← Redis + Circuit Breaker (sony/gobreaker)

[/api/v1/]
  Timeout             ← request timeout configurável
  AcceptJSON          ← Accept: application/json obrigatório
  AllowContentType    ← Content-Type: application/json obrigatório
  CSRF                ← HMAC-SHA256 double-submit (opcional)
```

---

## Stack Tecnológica

| Categoria          | Biblioteca                                     |
|--------------------|------------------------------------------------|
| DI                 | `go.uber.org/fx`                               |
| Router             | `github.com/go-chi/chi/v5`                     |
| Config             | `github.com/spf13/viper` + `.env`              |
| Logger             | `log/slog` (stdlib) wrapper                    |
| Error handling     | `github.com/marcelofabianov/fault`             |
| Validação          | `github.com/go-playground/validator/v10`       |
| Database           | `github.com/jackc/pgx/v5`                      |
| Cache              | `github.com/redis/go-redis/v9`                 |
| Migrations         | `github.com/pressly/goose/v3`                  |
| JWT                | `github.com/golang-jwt/jwt/v5`                 |
| Crypto             | `golang.org/x/crypto` (Argon2id)               |
| Rate Limit         | `github.com/go-redis/redis_rate/v10`           |
| Circuit Breaker    | `github.com/sony/gobreaker`                    |
| Sanitização        | `github.com/microcosm-cc/bluemonday`           |
| UUID               | `github.com/google/uuid`                       |
| Swagger            | `github.com/swaggo/swag` + `http-swagger`      |
| Testes             | `github.com/stretchr/testify`                  |
| Testcontainers     | `github.com/testcontainers/testcontainers-go`  |
| Lint               | `golangci-lint`                                |
| Security scan      | `gosec`                                        |
| Format             | `gofumpt`                                      |

---

## Makefile — Targets

```makefile
make help              # lista todos os comandos disponíveis

# desenvolvimento
make run               # sobe a aplicação com docker-compose
make build             # compila o binário
make certs             # gera certificados TLS self-signed para dev

# banco de dados
make migrate-up        # aplica migrations (goose up)
make migrate-down      # reverte última migration (goose down)
make migrate-status    # status das migrations

# qualidade
make test              # unit + integration
make test-unit         # go test -race ./...
make test-integration  # testcontainers
make test-coverage     # relatório de cobertura HTML
make lint              # golangci-lint
make lint-fix          # golangci-lint --fix
make fmt               # gofumpt
make security          # gosec
make quality           # lint + security + test

# documentação
make docs              # swag init → docs/

# docker
make shell             # shell no container api
make logs              # logs do container api (follow)
```

---

## HTTPS / TLS

```bash
# gerar certificados para desenvolvimento
make certs
# → cria certs/server.crt e certs/server.key (RSA 2048, self-signed, 365 dias)

# .env para habilitar TLS
APP_HTTP_TLS_ENABLED=true
APP_HTTP_TLS_CERT_FILE=certs/server.crt
APP_HTTP_TLS_KEY_FILE=certs/server.key
APP_HTTP_TLS_HTTPS_ONLY=true
```

O servidor configura automaticamente:
- TLS mínimo: 1.2 / máximo: 1.3
- Cipher suites modernos (AES-GCM, ChaCha20)
- Curve preferences: X25519, P-256

---

## CI/CD — GitHub Actions

```yaml
# .github/workflows/ci.yml
jobs:
  ci:
    steps:
      - golangci-lint     # lint estático
      - gosec             # análise de segurança
      - go test -race     # testes com race detector
      - go build          # verifica compilação
```

---

## Roadmap

- [x] Planejamento e análise do `personal-todo`
- [ ] CLI scaffold (cobra, embed.FS)
- [ ] Generator core (text/template)
- [ ] Templates: config, pkg/logger, pkg/web, middlewares
- [ ] Templates: pkg/database, cache, token, crypto, validation, retry
- [ ] Templates: internal/di/ (PkgModule + AppModule)
- [ ] Templates: internal/hello/ (bounded context de exemplo)
- [ ] Template: cmd/api/main.go
- [ ] Templates: Makefile, Dockerfile, docker-compose, scripts/gen-certs.sh
- [ ] Template: .github/workflows/ci.yml + .golangci.yml
- [ ] Template: go.mod, .env.example, .gitignore, README.md
- [ ] Modo interativo (prompts quando flags omitidas)
- [ ] Testes do gerador
- [ ] README do CLI
