package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"a.epub":    "fake epub",
		"sub/b.txt": "fake txt",
		"c.md":      "fake md",
		"d.zip":     "ignored",
		"e.exe":     "ignored",
	}
	for rel, content := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(got), got)
	}
	found := map[string]bool{}
	for _, p := range got {
		found[FormatFromExt(p)] = true
	}
	if !found["epub"] || !found["txt"] || !found["md"] {
		t.Fatalf("missing format: %v", found)
	}
}

func TestScanEmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty, got %v", got)
	}
}

func TestScanNonexistentDir(t *testing.T) {
	_, err := Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error")
	}
}
