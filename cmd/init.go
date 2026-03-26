package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/marcelofabianov/gostart/internal/generator"
)

var (
	flagModule   string
	flagDB       string
	flagNoCache  bool
	flagNoDocker bool
	flagNoCI     bool
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Inicializa um novo projeto Go REST API",
	Long: `Gera uma estrutura completa de projeto Go REST API com:
  • Arquitetura modular por bounded contexts
  • Segurança como prioridade (TLS, CSRF, rate-limit, headers)
  • DI via go.uber.org/fx
  • Roteamento Chi v5 com middleware stack completo
  • Módulo hello world como ponto de partida`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := ""
		if len(args) == 1 {
			projectName = args[0]
		}

		interactive := !cmd.Flags().Changed("module")

		if interactive {
			if err := runInteractivePrompts(cmd, &projectName); err != nil {
				return err
			}
		}

		if projectName == "" {
			return fmt.Errorf("nome do projeto é obrigatório")
		}
		if flagModule == "" {
			return fmt.Errorf("--module é obrigatório")
		}

		serviceName := projectName
		if parts := strings.Split(flagModule, "/"); len(parts) > 0 {
			serviceName = parts[len(parts)-1]
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("obter diretório atual: %w", err)
		}
		outputDir := filepath.Join(cwd, projectName)

		opts := generator.ProjectOptions{
			ProjectName: projectName,
			ModuleName:  flagModule,
			ServiceName: serviceName,
			DB:          flagDB,
			NoCache:     flagNoCache,
			NoDocker:    flagNoDocker,
			NoCI:        flagNoCI,
			OutputDir:   outputDir,
		}

		g := generator.New(opts)

		fmt.Printf("\n🚀 Criando projeto %q...\n", projectName)

		if err := g.Generate(); err != nil {
			return fmt.Errorf("erro ao gerar projeto: %w", err)
		}

		fmt.Printf("✅ Projeto %q criado com sucesso em %s\n", projectName, outputDir)
		fmt.Println("\nPróximos passos:")
		fmt.Printf("  cd %s\n", projectName)
		fmt.Println("  cp .env.example .env")
		fmt.Println("  make certs   # gera certificados TLS para dev")
		fmt.Println("  make run     # sobe com docker-compose")
		return nil
	},
}

func runInteractivePrompts(cmd *cobra.Command, projectName *string) error {
	dbChoice := "postgres"
	includeCache := true
	includeDocker := true
	includeCI := true

	fields := []huh.Field{
		huh.NewInput().
			Title("Nome do projeto").
			Description("Ex: minha-api, user-service, payments").
			Placeholder("meu-projeto").
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("nome do projeto não pode ser vazio")
				}
				return nil
			}).
			Value(projectName),

		huh.NewInput().
			Title("Caminho do módulo Go").
			Description("Ex: github.com/acme/minha-api").
			Placeholder("github.com/user/projeto").
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("módulo não pode ser vazio")
				}
				if !strings.Contains(s, "/") {
					return fmt.Errorf("módulo deve seguir o padrão host/path (ex: github.com/user/repo)")
				}
				return nil
			}).
			Value(&flagModule),

		huh.NewSelect[string]().
			Title("Banco de dados").
			Options(
				huh.NewOption("PostgreSQL (pgx v5)", "postgres"),
				huh.NewOption("Nenhum", "none"),
			).
			Value(&dbChoice),

		huh.NewConfirm().
			Title("Incluir Redis/cache?").
			Value(&includeCache),

		huh.NewConfirm().
			Title("Incluir Docker (Dockerfile + docker-compose)?").
			Value(&includeDocker),

		huh.NewConfirm().
			Title("Incluir GitHub Actions CI?").
			Value(&includeCI),
	}

	form := huh.NewForm(huh.NewGroup(fields...)).
		WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return fmt.Errorf("prompt interativo cancelado: %w", err)
	}

	flagDB = dbChoice
	flagNoCache = !includeCache
	flagNoDocker = !includeDocker
	flagNoCI = !includeCI

	return nil
}

func init() {
	initCmd.Flags().StringVar(&flagModule, "module", "", "Caminho do módulo Go (ex: github.com/user/projeto)")
	initCmd.Flags().StringVar(&flagDB, "db", "postgres", "Banco de dados: postgres ou none")
	initCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "Omite Redis/cache")
	initCmd.Flags().BoolVar(&flagNoDocker, "no-docker", false, "Omite Dockerfile e docker-compose")
	initCmd.Flags().BoolVar(&flagNoCI, "no-ci", false, "Omite GitHub Actions")
}
