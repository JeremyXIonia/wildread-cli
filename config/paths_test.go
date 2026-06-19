package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePathsDefault(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	p, err := ResolvePaths("", "", "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	wantData := filepath.Join(home, ".cli-read")
	if p.DataDir != wantData {
		t.Fatalf("data dir: %q, want %q", p.DataDir, wantData)
	}
	if p.DBPath != filepath.Join(wantData, "novel-reader.db") {
		t.Fatalf("db path: %q", p.DBPath)
	}
	if p.DefaultBookDir != filepath.Join(wantData, ".book") {
		t.Fatalf("book dir: %q", p.DefaultBookDir)
	}
	if p.TempBookDir != "" {
		t.Fatalf("temp dir: %q", p.TempBookDir)
	}
}

func TestResolvePathsDataDirAndDBOverride(t *testing.T) {
	base := t.TempDir()
	db := filepath.Join(t.TempDir(), "custom.db")
	p, err := ResolvePaths(base, db, "")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if p.DataDir != filepath.Clean(base) {
		t.Fatalf("data dir: %q", p.DataDir)
	}
	if p.DBPath != filepath.Clean(db) {
		t.Fatalf("db: %q", p.DBPath)
	}
	if p.DefaultBookDir != filepath.Join(filepath.Clean(base), ".book") {
		t.Fatalf("default book dir: %q", p.DefaultBookDir)
	}
}

func TestNormalizePathExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	got, err := NormalizePath("~/Books")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	want := filepath.Join(home, "Books")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestNormalizePathRelativeBecomesAbsolute(t *testing.T) {
	got, err := NormalizePath("relative-books")
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("not absolute: %q", got)
	}
	if !strings.HasSuffix(got, string(filepath.Separator)+"relative-books") {
		t.Fatalf("unexpected abs path: %q", got)
	}
}

func TestNormalizePathRejectsEmpty(t *testing.T) {
	if _, err := NormalizePath("   "); err == nil {
		t.Fatal("expected error for empty path")
	}
}
