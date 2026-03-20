package harness

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// PackageImportRule defines import boundaries for a package directory.
type PackageImportRule struct {
	PackageDir          string
	ForbiddenPrefixes   []string
	ArchitectureDocPath string
}

// CheckPackageImportRules validates that package directories do not import
// forbidden internal packages.
func CheckPackageImportRules(repoRoot string, rules []PackageImportRule) error {
	for _, rule := range rules {
		if err := checkPackageImportRule(repoRoot, rule); err != nil {
			return err
		}
	}

	return nil
}

// CheckRepoArchitecture validates the repository's documented architecture rules.
func CheckRepoArchitecture(repoRoot string) error {
	return CheckPackageImportRules(repoRoot, []PackageImportRule{
		{
			PackageDir:          "parser",
			ForbiddenPrefixes:   []string{"github.com/euforic/matr/matr", "github.com/euforic/matr/internal/harness"},
			ArchitectureDocPath: "docs/ARCHITECTURE.md",
		},
		{
			PackageDir:          "internal/harness",
			ForbiddenPrefixes:   []string{"github.com/euforic/matr/matr", "github.com/euforic/matr/parser"},
			ArchitectureDocPath: "docs/ARCHITECTURE.md",
		},
	})
}

func checkPackageImportRule(repoRoot string, rule PackageImportRule) error {
	packageRoot := filepath.Join(repoRoot, rule.PackageDir)
	var files []string

	err := filepath.WalkDir(packageRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(files)
	fset := token.NewFileSet()

	for _, file := range files {
		node, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}

		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, "\"")
			for _, forbidden := range rule.ForbiddenPrefixes {
				if path == forbidden || strings.HasPrefix(path, forbidden+"/") {
					return fmt.Errorf(
						"%s must not import %s.\nwhy: this breaks the package boundary documented in %s.\nbad: keep runtime or repo-workflow logic out of %s.\ngood: move runtime concerns into matr/, keep parsing in parser/, and keep repo checks in internal/harness/.\nfix: move runtime or repository-workflow logic out of %s and depend on a narrower helper or data type instead.\nsource: %s",
						rule.PackageDir,
						path,
						rule.ArchitectureDocPath,
						rule.PackageDir+"/",
						rule.PackageDir+"/",
						rule.ArchitectureDocPath,
					)
				}
			}
		}
	}

	return nil
}
