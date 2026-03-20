package matr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetProjectHashChangesWhenLocalGoFilesChange(t *testing.T) {
	dir := t.TempDir()

	matrfilePath := filepath.Join(dir, "Matrfile.go")
	helperPath := filepath.Join(dir, "helper.go")

	if err := os.WriteFile(matrfilePath, []byte("//go:build matr\n\npackage main\n"), 0644); err != nil {
		t.Fatalf("write matrfile: %v", err)
	}
	if err := os.WriteFile(helperPath, []byte("package main\n\nconst value = 1\n"), 0644); err != nil {
		t.Fatalf("write helper: %v", err)
	}

	initialHash, err := getProjectHash(matrfilePath)
	if err != nil {
		t.Fatalf("initial hash: %v", err)
	}

	if err := os.WriteFile(helperPath, []byte("package main\n\nconst value = 2\n"), 0644); err != nil {
		t.Fatalf("rewrite helper: %v", err)
	}

	updatedHash, err := getProjectHash(matrfilePath)
	if err != nil {
		t.Fatalf("updated hash: %v", err)
	}

	if string(initialHash) == string(updatedHash) {
		t.Fatalf("expected project hash to change when helper.go changed")
	}
}
