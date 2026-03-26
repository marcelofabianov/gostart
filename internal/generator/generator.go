package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Generator struct {
	opts ProjectOptions
}

func New(opts ProjectOptions) *Generator {
	return &Generator{opts: opts}
}

func (g *Generator) Generate() error {
	if err := os.MkdirAll(g.opts.OutputDir, 0755); err != nil {
		return fmt.Errorf("criar diretório: %w", err)
	}

	return fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel("templates", path)
		if rel == "." {
			return nil
		}

		destPath := filepath.Join(g.opts.OutputDir, rel)

		if strings.HasSuffix(destPath, ".tmpl") {
			destPath = strings.TrimSuffix(destPath, ".tmpl")
		}

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		if g.opts.NoDocker && (strings.Contains(path, "Dockerfile") || strings.Contains(path, "docker-compose")) {
			return nil
		}
		if g.opts.NoCI && strings.Contains(path, ".github") {
			return nil
		}
		if g.opts.NoCache && strings.Contains(path, "pkg/cache") {
			return nil
		}
		if g.opts.DB == "none" && strings.Contains(path, "pkg/database") {
			return nil
		}

		return g.renderFile(path, destPath)
	})
}

func (g *Generator) renderFile(srcPath, destPath string) error {
	content, err := templateFS.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("ler template %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("criar diretório para %s: %w", destPath, err)
	}

	if !strings.HasSuffix(srcPath, ".tmpl") {
		return os.WriteFile(destPath, content, 0644)
	}

	tmpl, err := template.New(filepath.Base(srcPath)).
		Delims("[[", "]]").
		Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsear template %s: %w", srcPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, g.opts); err != nil {
		return fmt.Errorf("executar template %s: %w", srcPath, err)
	}

	mode := fs.FileMode(0644)
	if strings.HasSuffix(destPath, ".sh") {
		mode = 0755
	}

	return os.WriteFile(destPath, buf.Bytes(), mode)
}
