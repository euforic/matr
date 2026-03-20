package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckRepoArchitecture(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", ".."))

	if err := CheckRepoArchitecture(repoRoot); err != nil {
		t.Fatalf("unexpected architecture violation: %v", err)
	}
}

func TestCheckPackageImportRules_AllowsDeclaredBoundaries(t *testing.T) {
	root := t.TempDir()

	writeGoFile(t, root, "parser/parser.go", "package parser\n")
	writeGoFile(t, root, "matr/runtime.go", "package matr\n\nimport _ \"github.com/euforic/matr/parser\"\n")
	writeGoFile(t, root, "internal/harness/check.go", "package harness\n")

	rules := []PackageImportRule{
		{
			PackageDir:         "parser",
			ForbiddenPrefixes:  []string{"github.com/euforic/matr/matr", "github.com/euforic/matr/internal/harness"},
			ArchitectureDocPath: "docs/ARCHITECTURE.md",
		},
		{
			PackageDir:         "internal/harness",
			ForbiddenPrefixes:  []string{"github.com/euforic/matr/matr", "github.com/euforic/matr/parser"},
			ArchitectureDocPath: "docs/ARCHITECTURE.md",
		},
	}

	if err := CheckPackageImportRules(root, rules); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckPackageImportRules_ReportsViolationWithFixGuidance(t *testing.T) {
	root := t.TempDir()

	writeGoFile(t, root, "parser/parser.go", "package parser\n\nimport _ \"github.com/euforic/matr/matr\"\n")

	err := CheckPackageImportRules(root, []PackageImportRule{
		{
			PackageDir:         "parser",
			ForbiddenPrefixes:  []string{"github.com/euforic/matr/matr", "github.com/euforic/matr/internal/harness"},
			ArchitectureDocPath: "docs/ARCHITECTURE.md",
		},
	})
	if err == nil {
		t.Fatalf("expected violation")
	}

	message := err.Error()
	for _, want := range []string{
		"parser must not import github.com/euforic/matr/matr",
		"move runtime or repository-workflow logic out of parser/",
		"docs/ARCHITECTURE.md",
		"bad:",
		"good:",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected %q in error message, got %q", want, message)
		}
	}
}

func writeGoFile(t *testing.T, root, relPath, content string) {
	t.Helper()

	fullPath := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
