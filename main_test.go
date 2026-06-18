package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuanchong/cli-read/models"
	"github.com/xuanchong/cli-read/store"
)

func newRootTestStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestEnsureDefaultLibraryDirCreatesWhenEmpty(t *testing.T) {
	st := newRootTestStore(t)
	defaultDir := filepath.Join(t.TempDir(), ".book")

	created, err := ensureDefaultLibraryDir(st, defaultDir)
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if !created {
		t.Fatal("expected created=true")
	}
	dirs, err := st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list dirs: %v", err)
	}
	if len(dirs) != 1 || dirs[0].Path != defaultDir || !dirs[0].IsDefault {
		t.Fatalf("dirs: %+v", dirs)
	}
}

func TestConfiguredScanDirsIncludesTemporaryDirWithoutPersisting(t *testing.T) {
	st := newRootTestStore(t)
	managed := filepath.Join(t.TempDir(), "managed")
	temp := filepath.Join(t.TempDir(), "temp")
	if _, err := st.AddLibraryDir(managed, true); err != nil {
		t.Fatalf("add managed: %v", err)
	}

	dirs, err := configuredScanDirs(st, temp)
	if err != nil {
		t.Fatalf("configured: %v", err)
	}
	if len(dirs) != 2 || dirs[0] != managed || dirs[1] != temp {
		t.Fatalf("dirs: %+v", dirs)
	}

	stored, err := st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(stored) != 1 || stored[0].Path != managed {
		t.Fatalf("stored dirs: %+v", stored)
	}
}

func TestRefreshBooksScansMultipleDirs(t *testing.T) {
	st := newRootTestStore(t)
	dirA := t.TempDir()
	dirB := t.TempDir()
	writeTestBook(t, filepath.Join(dirA, "a.txt"), "A")
	writeTestBook(t, filepath.Join(dirB, "b.txt"), "B")

	books, scanErrs, err := refreshBooks(st, []string{dirA, dirB})
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(scanErrs) != 0 {
		t.Fatalf("scan errors: %+v", scanErrs)
	}
	if len(books) != 2 {
		t.Fatalf("books: %+v", books)
	}
}

func TestSyncBooksPrunesMissingFromFullScan(t *testing.T) {
	st := newRootTestStore(t)
	dir := t.TempDir()
	bookPath := filepath.Join(dir, "a.txt")
	writeTestBook(t, bookPath, "A")

	if _, _, err := refreshBooks(st, []string{dir}); err != nil {
		t.Fatalf("first refresh: %v", err)
	}
	if err := os.Remove(bookPath); err != nil {
		t.Fatalf("remove book: %v", err)
	}
	books, _, err := refreshBooks(st, []string{dir})
	if err != nil {
		t.Fatalf("second refresh: %v", err)
	}
	if len(books) != 0 {
		t.Fatalf("books after prune: %+v", books)
	}
}

func writeTestBook(t *testing.T, path, title string) {
	t.Helper()
	content := title + "\n\n第一章\n内容"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write book: %v", err)
	}
}

var _ = models.Book{}
