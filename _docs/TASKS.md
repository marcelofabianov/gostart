# Checklist de Tarefas — `gostart`

> Acompanhamento da implementação do CLI de scaffolding para APIs Go REST.
> Ordem de execução respeita as dependências entre tarefas.

---

## Fase 1 — CLI Base

- [x] **`cli-scaffold`** — Inicializar módulo Go + instalar Cobra/Viper
  - `go mod init github.com/marcelofabianov/gostart`
  - Criar `main.go`, `cmd/root.go`, `cmd/init.go`
  - Flags: `--module`, `--db`, `--no-cache`, `--no-docker`, `--no-ci`

- [ ] **`generator-core`** — Criar `internal/generator/`
  - `generator.go` — orquestra a criação de arquivos via `embed.FS` + `text/template`
  - `options.go` — struct `ProjectOptions` com todas as configurações
  - `embed.go` — `//go:embed templates/**` embute tudo no binário

---

## Fase 2 — Templates: Fundação

- [ ] **`tpl-gomod`** — `go.mod` + `.gitignore` + `.env.example`
  - `go.mod` com todas as dependências (fault, wisp, fx, chi, pgx, viper, goose, validator, jwt, redis, testify...)
  - `.env.example` com todos os `APP_*` vars documentados e com valores padrão
  - `.gitignore` padrão Go

- [ ] **`tpl-config`** — `config/config.go` + `config/defaults.go`
  - Structs: `Config`, `GeneralConfig`, `HTTPConfig`, `TLSConfig`, `DatabaseConfig`, `RedisConfig`, `JWTConfig`, `AuthConfig`, `EmailConfig`...
  - `Load()` via Viper lendo `.env` + env vars (`APP_*`)
  - `setDefaults()` com valores seguros de produção

- [ ] **`tpl-migrations`** — `db/migrations/` + goose
  - Pasta `db/migrations/` com `.gitkeep`
  - Migration inicial `00001_init.sql` (esqueleto)
  - `MigrationsConfig` no `config.go`

---

## Fase 3 — Templates: Pacotes de Infraestrutura

- [ ] **`tpl-pkg-logger`** — `pkg/logger/`
  - `logger.go` — wrapper `log/slog`: `Debug`, `Info`, `Warn`, `Error`, `With`, `WithGroup`
  - `config.go` — `LogLevel`, `LogFormat`, `Config`

- [ ] **`tpl-pkg-infra`** — `pkg/database`, `pkg/cache`, `pkg/token`, `pkg/crypto`, `pkg/validation`, `pkg/retry`
  - `pkg/database/postgres.go` — pgx v5 + pool + retry + health check
  - `pkg/cache/cache.go` — go-redis v9 + pool + health check
  - `pkg/token/token.go` — JWT access + refresh (golang-jwt/jwt v5)
  - `pkg/crypto/hasher.go` — Argon2id hasher (hash + verify)
  - `pkg/validation/validator.go` — go-playground/validator + fault + redact sensível
  - `pkg/retry/retry.go` + `backoff_strategy.go` — exponential backoff com jitter

---

## Fase 4 — Templates: Web Layer

- [ ] **`tpl-pkg-web`** — `pkg/web/` (server, health, response, router, context)
  - `server.go` — HTTP Server + TLS 1.2/1.3 + graceful shutdown
  - `health.go` — `/health` (liveness) + `/health/ready` (readiness com checkers)
  - `response.go` — `Success`, `Error`, `Created`, `NoContent`, `BadRequest`...
  - `router.go` — `interface Router { RegisterRoutes(chi.Router) }`
  - `context.go` — `GetLogger`, `GetRequestID`, `SetUserID`, `GetUserID`...

- [ ] **`tpl-pkg-middleware`** — `pkg/web/middleware/` (14 arquivos)
  - `security_headers.go` — CSP, HSTS, X-Frame-Options, Referrer-Policy...
  - `https_only.go` — rejeita HTTP quando TLS habilitado
  - `csrf.go` — HMAC-SHA256 double-submit cookie
  - `rate_limit.go` — Redis rate limit + Sony Circuit Breaker
  - `security_logger.go` — log de eventos de segurança (CSRF, rate limit, IP spoofing)
  - `logger.go` — request logging com latência e status
  - `recovery.go` — panic recovery com stack trace
  - `request_id.go` — X-Request-ID header
  - `real_ip.go` — extração de IP real via trusted proxies
  - `cors.go` — CORS configurável por origem/método/header
  - `timeout.go` — request timeout configurável
  - `accept.go` — `Accept: application/json` enforcement
  - `sanitize.go` — sanitização de input (bluemonday)
  - `request_size.go` — limite de tamanho do body

- [ ] **`tpl-pkg-chi`** — `pkg/web/chi/router.go`
  - `NewRouter()` com a stack completa na ordem correta:
    ```
    Recovery → RequestID → RealIP → Logger → SecurityHeaders →
    HTTPSOnly → CORS → RequestSize → Compression → RateLimit →
    /ping → /health → /health/ready →
    /api/v1/ → Timeout → AcceptJSON → AllowContentType → CSRF → [rotas]
    ```

---

## Fase 5 — Templates: DI e Main

- [ ] **`tpl-di`** — `internal/di/pkg.go` + `internal/di/app.go`
  - `pkg.go` — `PkgModule`: `ProvideConfig`, `ProvideLogger`, `ProvideDatabase`, `ProvideCache`, `ProvideValidation`, `ProvideCrypto`...
  - `app.go` — `AppModule`: `ProvideRouter`, `ProvideServer`, `AsRouter()`, `AsHealthChecker()`, `DatabaseHealthChecker`, `CacheHealthChecker`

- [ ] **`tpl-main`** — `cmd/api/main.go`
  - `fx.New(di.PkgModule, di.AppModule, di.HelloModule, fx.Invoke(func(*web.Server) {})).Run()`
  - Anotações Swagger (`@title`, `@version`, `@BasePath /api/v1`, `@securityDefinitions.apikey Bearer`)

---

## Fase 6 — Templates: Módulo Hello World (Bounded Context)

- [ ] **`tpl-domain`** — `internal/hello/` — módulo de exemplo completo
  - `domain/entity.go` — entidade `Hello` com `ID uuid`, `Message string`, `CreatedAt`
  - `port/port.go` — interface `SayHelloPort`
  - `usecase/usecase.go` — `SayHelloUseCase` implementa `SayHelloPort`
  - `handler/handler.go` — `HelloHandler` implementa `web.Router`
    - `GET /api/v1/hello` → `{"message": "Hello, World!", "module": "hello"}`
  - `storage/.gitkeep` — placeholder para repositório
  - `publisher/.gitkeep` — placeholder para publisher de eventos
  - `internal/di/hello.go` — `HelloModule` com `AsRouter(handler.NewHelloHandler)`

---

## Fase 7 — Templates: Infraestrutura de Projeto

- [ ] **`tpl-infra`** — Makefile + Dockerfile + docker-compose + scripts
  - `Makefile` com targets: `help`, `run`, `build`, `certs`, `migrate-up`, `migrate-down`, `migrate-status`, `test`, `test-unit`, `test-integration`, `test-coverage`, `lint`, `lint-fix`, `fmt`, `security`, `quality`, `docs`, `shell`, `logs`
  - `docker/Dockerfile` — multi-stage: `builder` (golang) + runtime (`distroless/static`)
  - `docker-compose.yml` — `postgres` + `redis` + `app` com healthchecks
  - `scripts/gen-certs.sh` — gera `certs/server.crt` + `certs/server.key` (RSA 2048, self-signed)
  - `.golangci.yml` — configuração do linter

- [ ] **`tpl-ci`** — `.github/workflows/ci.yml`
  - Jobs: `golangci-lint` + `gosec` + `go test -race ./...` + `go build ./...`
  - Cache de módulos Go
  - `ubuntu-latest`, Go versão do `go.mod`

---

## Fase 8 — Interatividade e Qualidade

- [ ] **`interactive`** — Modo interativo com prompts
  - Quando flags omitidas: perguntar `module path`, `project name`, `db`, `cache`, `docker`, `ci`
  - Usar `github.com/charmbracelet/huh` (bubbletea) ou `AlecAivazis/survey`
  - Mostrar resumo antes de gerar e pedir confirmação

- [ ] **`generator-tests`** — Testes do gerador
  - Verificar estrutura de diretórios gerada
  - Verificar existência de arquivos críticos
  - Verificar que `go.mod` contém o `--module` correto
  - Verificar que o projeto gerado compila com `go build ./...`

- [ ] **`readme-cli`** — `README.md` do CLI `gostart`
  - Instalação, uso, flags, estrutura gerada, stack, exemplos
  - Badge CI, Go version, license

---

## Resumo de Dependências

```
cli-scaffold
  └── generator-core
        └── tpl-gomod
              ├── tpl-config
              │     ├── tpl-migrations
              │     ├── tpl-pkg-logger
              │     │     ├── tpl-pkg-web
              │     │     │     └── tpl-pkg-middleware
              │     │     │           └── tpl-pkg-chi
              │     │     │                 └── tpl-di ──────────┐
              │     │     └── tpl-pkg-infra                      │
              │     │           └── tpl-di ─────────────────────►│
              │     └── (config pronto)                          │
              └── (gomod pronto)                                  │
                                                         tpl-main (dep: tpl-di)
                                                           └── tpl-infra
                                                                 └── tpl-ci
                                                                 └── generator-tests
                                                                       └── readme-cli

tpl-di → tpl-domain (internal/hello/)
tpl-main → interactive
```

---

## Status

| Fase | Tarefas | Concluídas |
|------|---------|-----------|
| 1 — CLI Base           | 2  | 1 |
| 2 — Fundação           | 3  | 0 |
| 3 — Infra Packages     | 2  | 0 |
| 4 — Web Layer          | 3  | 0 |
| 5 — DI e Main          | 2  | 0 |
| 6 — Hello World Module | 1  | 0 |
| 7 — Infra de Projeto   | 2  | 0 |
| 8 — Qualidade          | 3  | 0 |
| **Total**              | **18** | **1** |
