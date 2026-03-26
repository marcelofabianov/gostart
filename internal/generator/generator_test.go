package generator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/marcelofabianov/gostart/internal/generator"
)

func newTestOptions(t *testing.T, overrides ...func(*generator.ProjectOptions)) generator.ProjectOptions {
	t.Helper()
	opts := generator.ProjectOptions{
		ProjectName: "test-project",
		ModuleName:  "github.com/test/test-project",
		ServiceName: "test-project",
		DB:          "postgres",
		NoCache:     false,
		NoDocker:    false,
		NoCI:        false,
		OutputDir:   filepath.Join(t.TempDir(), "test-project"),
	}
	for _, fn := range overrides {
		fn(&opts)
	}
	return opts
}

// ProjectOptionsMock mirrors generator.ProjectOptions to avoid import cycle.
func TestGenerate_DefaultProject(t *testing.T) {
	opts := newTestOptions(t)
	g := generator.New(opts)

	if err := g.Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	requiredFiles := []string{
		"go.mod",
		"cmd/api/main.go",
		"config/config.go",
		"internal/di/app.go",
		"internal/di/pkg.go",
		"internal/di/hello.go",
		"internal/hello/domain/entity.go",
		"internal/hello/domain/errors.go",
		"internal/hello/domain/events.go",
		"internal/hello/handler/handler.go",
		"internal/hello/port/usecase.go",
		"internal/hello/port/repository.go",
		"internal/hello/usecase/usecase.go",
		"internal/hello/storage/memory_hello_repository.go",
		"internal/hello/publisher/.gitkeep",
		"pkg/web/server.go",
		"pkg/web/router.go",
		"pkg/web/health.go",
		"pkg/web/response.go",
		"pkg/web/context.go",
		"pkg/web/chi/router.go",
		"pkg/web/middleware/cors.go",
		"pkg/web/middleware/security_headers.go",
		"pkg/web/middleware/rate_limit.go",
		"pkg/web/middleware/csrf.go",
		"pkg/web/middleware/timeout.go",
		"pkg/database/postgres.go",
		"pkg/cache/cache.go",
		"pkg/logger/logger.go",
		"pkg/retry/retry.go",
		"pkg/retry/backoff_strategy.go",
		"pkg/token/token.go",
		"pkg/crypto/hasher.go",
		"pkg/validation/validator.go",
		"pkg/validation/config.go",
		"Makefile",
		"README.md",
		"scripts/gen-certs.sh",
		"docker/Dockerfile",
		"docker-compose.yml",
		"docs/docs.go",
		"db/migrations/00001_init.sql",
		".gitignore",
		".env.example",
		".golangci.yml",
		".github/workflows/ci.yml",
	}

	for _, rel := range requiredFiles {
		full := filepath.Join(opts.OutputDir, rel)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("arquivo esperado não encontrado: %s", rel)
		}
	}
}

func TestGenerate_ModuleSubstitution(t *testing.T) {
	opts := newTestOptions(t)
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	checkContains := func(relPath, want string) {
		t.Helper()
		full := filepath.Join(opts.OutputDir, relPath)
		data, err := os.ReadFile(full)
		if err != nil {
			t.Fatalf("ler %s: %v", relPath, err)
		}
		if !strings.Contains(string(data), want) {
			t.Errorf("%s: esperado conter %q", relPath, want)
		}
	}

	checkContains("go.mod", "module github.com/test/test-project")
	checkContains("cmd/api/main.go", "github.com/test/test-project/internal/di")
	checkContains("internal/di/hello.go", "github.com/test/test-project/internal/hello/handler")
	checkContains(".env.example", "APP_GENERAL_SERVICE_NAME=test-project")
	checkContains("Makefile", "test-project")
}

func TestGenerate_NoDocker(t *testing.T) {
	opts := newTestOptions(t, func(o *generator.ProjectOptions) {
		o.NoDocker = true
	})
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	absent := []string{"docker/Dockerfile", "docker-compose.yml"}
	for _, rel := range absent {
		full := filepath.Join(opts.OutputDir, rel)
		if _, err := os.Stat(full); !os.IsNotExist(err) {
			t.Errorf("arquivo %s não deveria existir com --no-docker", rel)
		}
	}

	// core files must still exist
	gomod := filepath.Join(opts.OutputDir, "go.mod")
	if _, err := os.Stat(gomod); os.IsNotExist(err) {
		t.Error("go.mod deve existir mesmo com --no-docker")
	}
}

func TestGenerate_NoCI(t *testing.T) {
	opts := newTestOptions(t, func(o *generator.ProjectOptions) {
		o.NoCI = true
	})
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	ci := filepath.Join(opts.OutputDir, ".github/workflows/ci.yml")
	if _, err := os.Stat(ci); !os.IsNotExist(err) {
		t.Error(".github/workflows/ci.yml não deveria existir com --no-ci")
	}
}

func TestGenerate_NoCache(t *testing.T) {
	opts := newTestOptions(t, func(o *generator.ProjectOptions) {
		o.NoCache = true
	})
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	cache := filepath.Join(opts.OutputDir, "pkg/cache/cache.go")
	if _, err := os.Stat(cache); !os.IsNotExist(err) {
		t.Error("pkg/cache/cache.go não deveria existir com --no-cache")
	}
}

func TestGenerate_DBNone(t *testing.T) {
	opts := newTestOptions(t, func(o *generator.ProjectOptions) {
		o.DB = "none"
	})
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	db := filepath.Join(opts.OutputDir, "pkg/database/postgres.go")
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Error("pkg/database/postgres.go não deveria existir com --db=none")
	}
}

func TestGenerate_ScriptPermissions(t *testing.T) {
	opts := newTestOptions(t)
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	script := filepath.Join(opts.OutputDir, "scripts/gen-certs.sh")
	info, err := os.Stat(script)
	if err != nil {
		t.Fatalf("scripts/gen-certs.sh não encontrado: %v", err)
	}

	mode := info.Mode()
	if mode&0100 == 0 {
		t.Errorf("scripts/gen-certs.sh deve ter permissão de execução, mode=%v", mode)
	}
}

func TestGenerate_NoDotfilesMissing(t *testing.T) {
	opts := newTestOptions(t)
	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	dotfiles := []string{".gitignore", ".env.example", ".golangci.yml"}
	for _, rel := range dotfiles {
		full := filepath.Join(opts.OutputDir, rel)
		if _, err := os.Stat(full); os.IsNotExist(err) {
			t.Errorf("dotfile %s não deveria ser omitido pelo embed.FS", rel)
		}
	}
}

func TestGenerate_OutputDirCreated(t *testing.T) {
	base := t.TempDir()
	opts := generator.ProjectOptions{
		ProjectName: "novo",
		ModuleName:  "github.com/x/novo",
		ServiceName: "novo",
		DB:          "postgres",
		OutputDir:   filepath.Join(base, "deep", "nested", "novo"),
	}

	if err := generator.New(opts).Generate(); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if _, err := os.Stat(opts.OutputDir); os.IsNotExist(err) {
		t.Error("OutputDir deve ser criado automaticamente")
	}
}
