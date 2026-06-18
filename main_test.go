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

func TestRefreshBooksPreservesExistingBooksWhenOneDirScanFails(t *testing.T) {
	st := newRootTestStore(t)
	dirOK := t.TempDir()
	dirFailed := t.TempDir()
	okPath := filepath.Join(dirOK, "ok.txt")
	failedPath := filepath.Join(dirFailed, "preserve.txt")
	writeTestBook(t, okPath, "OK")
	writeTestBook(t, failedPath, "Preserve")

	if _, _, err := refreshBooks(st, []string{dirOK, dirFailed}); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	if err := os.RemoveAll(dirFailed); err != nil {
		t.Fatalf("remove failed dir: %v", err)
	}

	books, scanErrs, err := refreshBooks(st, []string{dirOK, dirFailed})
	if err != nil {
		t.Fatalf("refresh with failed dir: %v", err)
	}
	if len(scanErrs) != 1 {
		t.Fatalf("scan errors: %+v", scanErrs)
	}
	if !hasBookPath(books, failedPath) {
		t.Fatalf("expected failed directory book %s to be preserved, books: %+v", failedPath, books)
	}
	if !hasBookPath(books, okPath) {
		t.Fatalf("expected successful directory book %s to remain, books: %+v", okPath, books)
	}
}

func TestRefreshBooksPrunesStaleBooksUnderSuccessfulRootWhenAnotherRootFails(t *testing.T) {
	st := newRootTestStore(t)
	dirOK := t.TempDir()
	dirFailed := t.TempDir()
	staleOKPath := filepath.Join(dirOK, "stale.txt")
	freshOKPath := filepath.Join(dirOK, "fresh.txt")
	failedPath := filepath.Join(dirFailed, "preserve.txt")
	writeTestBook(t, staleOKPath, "Stale")
	writeTestBook(t, failedPath, "Preserve")

	if _, _, err := refreshBooks(st, []string{dirOK, dirFailed}); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	if err := os.Remove(staleOKPath); err != nil {
		t.Fatalf("remove stale book: %v", err)
	}
	writeTestBook(t, freshOKPath, "Fresh")
	if err := os.RemoveAll(dirFailed); err != nil {
		t.Fatalf("remove failed dir: %v", err)
	}

	books, scanErrs, err := refreshBooks(st, []string{dirOK, dirFailed})
	if err != nil {
		t.Fatalf("refresh with failed dir: %v", err)
	}
	if len(scanErrs) != 1 {
		t.Fatalf("scan errors: %+v", scanErrs)
	}
	if hasBookPath(books, staleOKPath) {
		t.Fatalf("expected stale successful-root book %s to be pruned, books: %+v", staleOKPath, books)
	}
	if !hasBookPath(books, freshOKPath) {
		t.Fatalf("expected fresh successful-root book %s to be upserted, books: %+v", freshOKPath, books)
	}
	if !hasBookPath(books, failedPath) {
		t.Fatalf("expected failed-root book %s to be preserved, books: %+v", failedPath, books)
	}
}

func TestRefreshBooksPrunesManagedStaleBooksWhenTemporaryRootFails(t *testing.T) {
	st := newRootTestStore(t)
	managedDir := t.TempDir()
	tempDir := t.TempDir()
	staleManagedPath := filepath.Join(managedDir, "stale.txt")
	freshManagedPath := filepath.Join(managedDir, "fresh.txt")
	writeTestBook(t, staleManagedPath, "Stale")

	if _, _, err := refreshBooks(st, []string{managedDir}); err != nil {
		t.Fatalf("initial refresh: %v", err)
	}
	if err := os.Remove(staleManagedPath); err != nil {
		t.Fatalf("remove stale managed book: %v", err)
	}
	writeTestBook(t, freshManagedPath, "Fresh")
	if err := os.RemoveAll(tempDir); err != nil {
		t.Fatalf("remove temp dir: %v", err)
	}

	books, scanErrs, err := refreshBooks(st, []string{managedDir, tempDir})
	if err != nil {
		t.Fatalf("refresh with failed temp dir: %v", err)
	}
	if len(scanErrs) != 1 {
		t.Fatalf("scan errors: %+v", scanErrs)
	}
	if hasBookPath(books, staleManagedPath) {
		t.Fatalf("expected stale managed book %s to be pruned, books: %+v", staleManagedPath, books)
	}
	if !hasBookPath(books, freshManagedPath) {
		t.Fatalf("expected fresh managed book %s to be upserted, books: %+v", freshManagedPath, books)
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

func TestStartupStatusSurfacesScanWarningsWhenDefaultCreated(t *testing.T) {
	status := startupStatus(2, true, []error{os.ErrNotExist}, filepath.Join(t.TempDir(), ".book"), "")
	want := "已扫描 2 本书，1 个目录扫描失败"
	if status != want {
		t.Fatalf("status = %q, want %q", status, want)
	}
}

func hasBookPath(books []models.Book, path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	for _, book := range books {
		if book.FilePath == abs {
			return true
		}
	}
	return false
}

func writeTestBook(t *testing.T, path, title string) {
	t.Helper()
	content := title + "\n\n第一章\n内容"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write book: %v", err)
	}
}

var _ = models.Book{}
