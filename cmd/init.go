package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		if flagModule == "" {
			return fmt.Errorf("--module é obrigatório")
		}

		serviceName := projectName
		if parts := strings.Split(flagModule, "/"); len(parts) > 0 {
			serviceName = parts[len(parts)-1]
		}

		opts := generator.ProjectOptions{
			ProjectName: projectName,
			ModuleName:  flagModule,
			ServiceName: serviceName,
			DB:          flagDB,
			NoCache:     flagNoCache,
			NoDocker:    flagNoDocker,
			NoCI:        flagNoCI,
			OutputDir:   filepath.Join(".", projectName),
		}

		g := generator.New(opts)

		fmt.Printf("🚀 Criando projeto %q...\n", projectName)

		if err := g.Generate(); err != nil {
			return fmt.Errorf("erro ao gerar projeto: %w", err)
		}

		fmt.Printf("✅ Projeto %q criado com sucesso em ./%s\n", projectName, projectName)
		fmt.Println("\nPróximos passos:")
		fmt.Printf("  cd %s\n", projectName)
		fmt.Println("  cp .env.example .env")
		fmt.Println("  make certs   # gera certificados TLS para dev")
		fmt.Println("  make run     # sobe com docker-compose")
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&flagModule, "module", "", "Caminho do módulo Go (obrigatório, ex: github.com/user/projeto)")
	initCmd.Flags().StringVar(&flagDB, "db", "postgres", "Banco de dados: postgres ou none")
	initCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "Omite Redis/cache")
	initCmd.Flags().BoolVar(&flagNoDocker, "no-docker", false, "Omite Dockerfile e docker-compose")
	initCmd.Flags().BoolVar(&flagNoCI, "no-ci", false, "Omite GitHub Actions")
	_ = initCmd.MarkFlagRequired("module")
}
