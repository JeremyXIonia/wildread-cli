# Storage Directory Management Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add stable app data storage under the user's home directory, persist managed book directories in SQLite, and provide a bookshelf-accessible TUI for viewing, adding, deleting, and rescanning book directories.

**Architecture:** Add a small `config` package for path resolution, extend `models` and `store` with `LibraryDir` persistence, and keep multi-directory scanning orchestration in `main.go`. Add an independent `app.DirectoryManagerModel` that emits messages for add/delete/rescan/exit; the root model owns store mutations and bookshelf refreshes.

**Tech Stack:** Go 1.25+, Bubble Tea, Bubbles list/textinput, SQLite via `modernc.org/sqlite`, standard `os/user`, `path/filepath`, and `strings` packages.

## Global Constraints

- Default app data directory: `~/.cli-read`.
- Default database path: `~/.cli-read/novel-reader.db`.
- Default managed book directory: `~/.cli-read/.book`.
- `--data-dir <dir>` sets both default database and default book directory unless `--db` overrides the database file.
- `--db <file>` remains an advanced database-file override.
- `--dir <dir>` remains as a one-session temporary scan directory and does not write to `library_dirs`.
- Path normalization expands `~`, converts relative paths to absolute paths, and stores cleaned absolute paths.
- Deleting a managed directory deletes book rows under that directory and relies on foreign-key cascade for progress/bookmarks; original files on disk are never deleted.
- Directory changes refresh the bookshelf immediately.
- If no managed directories exist, recreate and re-add the default book directory.
- Product docs must document Go 1.25+ and the new default storage paths.

---

## File Structure

- Create: `config/paths.go` — resolve app data directory, database path, default book directory, and normalized user paths.
- Create: `config/paths_test.go` — tests for default, override, `~`, and relative path normalization.
- Modify: `models/book.go` — add `LibraryDir` model.
- Modify: `store/schema.sql` — add `library_dirs` table.
- Create: `store/library_dirs.go` — CRUD for library dirs and destructive cleanup for books under a dir.
- Modify: `store/store_test.go` — add store tests for library dirs and cascade cleanup.
- Modify: `main.go` — parse `--data-dir`, ensure default dirs, scan multiple dirs, handle directory manager messages, refresh bookshelf.
- Create: `main_test.go` — test startup helpers and multi-directory scan/sync behavior.
- Modify: `app/messages.go` — add directory manager/root messages.
- Create: `app/directory_manager.go` — independent TUI model for directory management.
- Create: `app/directory_manager_test.go` — tests for view, add, create confirmation, delete confirmation, and exit messages.
- Modify: `app/bookshelf.go` — add `D` shortcut to open directory manager.
- Modify: `ui/keys.go` — add a `ManageDirs` key binding.
- Modify: `README.md` and `CLAUDE.md` — document new data directory, directory manager shortcut, and flag semantics.

---

### Task 1: Path Resolution Package

**Files:**
- Create: `config/paths.go`
- Create: `config/paths_test.go`

**Interfaces:**
- Produces:
  - `type Paths struct { DataDir string; DBPath string; DefaultBookDir string; TempBookDir string }`
  - `func ResolvePaths(dataDirFlag, dbFlag, tempDirFlag string) (Paths, error)`
  - `func NormalizePath(path string) (string, error)`

- [ ] **Step 1: Write failing path tests**

Create `config/paths_test.go`:

```go
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
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
go test ./config -v
```

Expected: FAIL because package `config` or functions are undefined.

- [ ] **Step 3: Implement path resolution**

Create `config/paths.go`:

```go
package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	DefaultDataDirName = ".cli-read"
	DefaultDBFileName  = "novel-reader.db"
	DefaultBookDirName = ".book"
)

type Paths struct {
	DataDir        string
	DBPath         string
	DefaultBookDir string
	TempBookDir    string
}

func ResolvePaths(dataDirFlag, dbFlag, tempDirFlag string) (Paths, error) {
	dataDir := dataDirFlag
	if strings.TrimSpace(dataDir) == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}
		dataDir = filepath.Join(home, DefaultDataDirName)
	}

	dataDir, err := NormalizePath(dataDir)
	if err != nil {
		return Paths{}, err
	}

	dbPath := filepath.Join(dataDir, DefaultDBFileName)
	if strings.TrimSpace(dbFlag) != "" {
		dbPath, err = NormalizePath(dbFlag)
		if err != nil {
			return Paths{}, err
		}
	}

	var tempDir string
	if strings.TrimSpace(tempDirFlag) != "" {
		tempDir, err = NormalizePath(tempDirFlag)
		if err != nil {
			return Paths{}, err
		}
	}

	return Paths{
		DataDir:        dataDir,
		DBPath:         dbPath,
		DefaultBookDir: filepath.Join(dataDir, DefaultBookDirName),
		TempBookDir:    tempDir,
	}, nil
}

func NormalizePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", errors.New("empty path")
	}
	if path == "~" || strings.HasPrefix(path, "~"+string(filepath.Separator)) || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			path = home
		} else {
			path = filepath.Join(home, path[2:])
		}
	}
	if !filepath.IsAbs(path) {
		abs, err := filepath.Abs(path)
		if err != nil {
			return "", err
		}
		path = abs
	}
	return filepath.Clean(path), nil
}
```

- [ ] **Step 4: Run path tests**

Run:

```bash
go test ./config -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add config/paths.go config/paths_test.go
git commit -m "feat: 解析应用数据目录"
```

---

### Task 2: Persist Managed Book Directories

**Files:**
- Modify: `models/book.go`
- Modify: `store/schema.sql`
- Create: `store/library_dirs.go`
- Modify: `store/store_test.go`

**Interfaces:**
- Consumes: `models.Book`, existing `Store`.
- Produces:
  - `models.LibraryDir`
  - `func (s *Store) ListLibraryDirs() ([]models.LibraryDir, error)`
  - `func (s *Store) AddLibraryDir(path string, isDefault bool) (int64, error)`
  - `func (s *Store) DeleteLibraryDir(id int64) error`
  - `func (s *Store) DeleteBooksUnderDir(dir string) error`

- [ ] **Step 1: Add failing store tests**

Append to `store/store_test.go`:

```go
func TestLibraryDirsCRUD(t *testing.T) {
	st := newTestStore(t)

	id, err := st.AddLibraryDir("/tmp/books", true)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	dirs, err := st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("dirs len: %d", len(dirs))
	}
	if dirs[0].ID != id || dirs[0].Path != "/tmp/books" || !dirs[0].IsDefault {
		t.Fatalf("dir: %+v", dirs[0])
	}

	if _, err := st.AddLibraryDir("/tmp/books", false); err == nil {
		t.Fatal("expected duplicate error")
	}

	if err := st.DeleteLibraryDir(id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	dirs, err = st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(dirs) != 0 {
		t.Fatalf("dirs after delete: %+v", dirs)
	}
}

func TestDeleteBooksUnderDirCascadesProgressAndBookmarks(t *testing.T) {
	st := newTestStore(t)

	keepID, err := st.UpsertBook(models.Book{FilePath: "/books-keep/a.txt", Title: "keep", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert keep: %v", err)
	}
	deleteID, err := st.UpsertBook(models.Book{FilePath: "/books-delete/sub/b.txt", Title: "delete", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert delete: %v", err)
	}
	if err := st.SaveProgress(models.ReadingProgress{BookID: deleteID, Chapter: 1, Page: 2}); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if _, err := st.AddBookmark(models.Bookmark{BookID: deleteID, Chapter: 0, Page: 0, Label: "mark"}); err != nil {
		t.Fatalf("bookmark: %v", err)
	}

	if err := st.DeleteBooksUnderDir("/books-delete"); err != nil {
		t.Fatalf("delete under dir: %v", err)
	}

	books, err := st.ListBooks()
	if err != nil {
		t.Fatalf("list books: %v", err)
	}
	if len(books) != 1 || books[0].ID != keepID {
		t.Fatalf("books: %+v", books)
	}
	if _, err := st.GetBook(deleteID); err == nil {
		t.Fatal("deleted book still exists")
	}
	progress, err := st.GetProgress(deleteID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if progress.Chapter != 0 || progress.Page != 0 {
		t.Fatalf("progress not cascaded: %+v", progress)
	}
	marks, err := st.ListBookmarks(deleteID)
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}
	if len(marks) != 0 {
		t.Fatalf("bookmarks not cascaded: %+v", marks)
	}
}
```

- [ ] **Step 2: Run failing store tests**

Run:

```bash
go test ./store -run 'TestLibraryDirsCRUD|TestDeleteBooksUnderDirCascadesProgressAndBookmarks' -v
```

Expected: FAIL because `AddLibraryDir`, `ListLibraryDirs`, `DeleteLibraryDir`, and `DeleteBooksUnderDir` are undefined.

- [ ] **Step 3: Add model**

Append to `models/book.go`:

```go
type LibraryDir struct {
	ID        int64
	Path      string
	IsDefault bool
	CreatedAt string
}
```

- [ ] **Step 4: Add schema**

Append to `store/schema.sql`:

```sql

CREATE TABLE IF NOT EXISTS library_dirs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    path        TEXT NOT NULL UNIQUE,
    is_default  INTEGER NOT NULL DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

- [ ] **Step 5: Implement store API**

Create `store/library_dirs.go`:

```go
package store

import (
	"path/filepath"
	"strings"

	"github.com/xuanchong/cli-read/models"
)

func (s *Store) ListLibraryDirs() ([]models.LibraryDir, error) {
	rows, err := s.db.Query(`SELECT id, path, is_default, created_at FROM library_dirs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.LibraryDir
	for rows.Next() {
		var d models.LibraryDir
		var isDefault int
		if err := rows.Scan(&d.ID, &d.Path, &isDefault, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.IsDefault = isDefault != 0
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) AddLibraryDir(path string, isDefault bool) (int64, error) {
	defaultInt := 0
	if isDefault {
		defaultInt = 1
	}
	res, err := s.db.Exec(`INSERT INTO library_dirs (path, is_default) VALUES (?, ?)`, path, defaultInt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteLibraryDir(id int64) error {
	_, err := s.db.Exec(`DELETE FROM library_dirs WHERE id=?`, id)
	return err
}

func (s *Store) DeleteBooksUnderDir(dir string) error {
	dir = filepath.Clean(dir)
	prefix := dir + string(filepath.Separator) + "%"
	if strings.HasSuffix(dir, string(filepath.Separator)) {
		prefix = dir + "%"
	}
	_, err := s.db.Exec(`DELETE FROM books WHERE file_path = ? OR file_path LIKE ?`, dir, prefix)
	return err
}
```

- [ ] **Step 6: Run store tests**

Run:

```bash
go test ./store -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add models/book.go store/schema.sql store/library_dirs.go store/store_test.go
git commit -m "feat: 存储书籍目录列表"
```

---

### Task 3: Startup Defaults and Multi-Directory Sync

**Files:**
- Modify: `main.go`
- Create: `main_test.go`

**Interfaces:**
- Consumes:
  - `config.ResolvePaths(dataDirFlag, dbFlag, tempDirFlag string) (config.Paths, error)`
  - `(*store.Store).ListLibraryDirs() ([]models.LibraryDir, error)`
  - `(*store.Store).AddLibraryDir(path string, isDefault bool) (int64, error)`
- Produces:
  - `func ensureDefaultLibraryDir(st *store.Store, defaultDir string) (bool, error)`
  - `func configuredScanDirs(st *store.Store, tempDir string) ([]string, error)`
  - `func scanAllDirs(dirs []string) ([]string, []error)`
  - `func refreshBooks(st *store.Store, dirs []string) ([]models.Book, []error, error)`

- [ ] **Step 1: Add startup helper tests**

Create `main_test.go`:

```go
package main

import (
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
	if err := st.DeleteBookFileForTest(bookPath); err == nil {
		t.Fatal("unexpected helper method exists")
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
```

Then immediately fix the intentional bad helper in the same file before running by replacing `TestSyncBooksPrunesMissingFromFullScan` with this version:

```go
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
```

Ensure imports are exactly:

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuanchong/cli-read/models"
	"github.com/xuanchong/cli-read/store"
)
```

- [ ] **Step 2: Run failing tests**

Run:

```bash
go test . -run 'TestEnsureDefaultLibraryDirCreatesWhenEmpty|TestConfiguredScanDirsIncludesTemporaryDirWithoutPersisting|TestRefreshBooksScansMultipleDirs|TestSyncBooksPrunesMissingFromFullScan' -v
```

Expected: FAIL because helper functions are undefined.

- [ ] **Step 3: Update main flags and path setup**

In `main.go`, add import:

```go
"github.com/xuanchong/cli-read/config"
```

Replace flag setup and initial directory/database setup with:

```go
dataDirFlag := flag.String("data-dir", "", "应用数据目录")
tempDirFlag := flag.String("dir", "", "临时书籍目录（本次扫描，不保存）")
dbPathFlag := flag.String("db", "", "SQLite 数据库路径")
flag.Parse()

paths, err := config.ResolvePaths(*dataDirFlag, *dbPathFlag, *tempDirFlag)
if err != nil {
	fmt.Fprintf(os.Stderr, "解析路径失败: %v\n", err)
	os.Exit(1)
}

if err := os.MkdirAll(paths.DataDir, 0755); err != nil {
	fmt.Fprintf(os.Stderr, "无法创建应用数据目录: %v\n", err)
	os.Exit(1)
}
if err := os.MkdirAll(paths.DefaultBookDir, 0755); err != nil {
	fmt.Fprintf(os.Stderr, "无法创建默认书籍目录: %v\n", err)
	os.Exit(1)
}
```

Open store with:

```go
st, err := store.Open(paths.DBPath)
```

- [ ] **Step 4: Add startup/refresh helpers**

Add below `main()` in `main.go`:

```go
func ensureDefaultLibraryDir(st *store.Store, defaultDir string) (bool, error) {
	dirs, err := st.ListLibraryDirs()
	if err != nil {
		return false, err
	}
	if len(dirs) > 0 {
		return false, nil
	}
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return false, err
	}
	_, err = st.AddLibraryDir(defaultDir, true)
	return err == nil, err
}

func configuredScanDirs(st *store.Store, tempDir string) ([]string, error) {
	libraryDirs, err := st.ListLibraryDirs()
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(libraryDirs)+1)
	for _, d := range libraryDirs {
		dirs = append(dirs, d.Path)
	}
	if tempDir != "" {
		dirs = append(dirs, tempDir)
	}
	return dirs, nil
}

func scanAllDirs(dirs []string) ([]string, []error) {
	seen := map[string]bool{}
	var paths []string
	var errs []error
	for _, dir := range dirs {
		scanned, err := parser.Scan(dir)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		for _, p := range scanned {
			abs, err := filepath.Abs(p)
			if err != nil {
				abs = p
			}
			if !seen[abs] {
				seen[abs] = true
				paths = append(paths, abs)
			}
		}
	}
	return paths, errs
}

func refreshBooks(st *store.Store, dirs []string) ([]models.Book, []error, error) {
	paths, scanErrs := scanAllDirs(dirs)
	if err := syncBooks(st, paths); err != nil {
		return nil, scanErrs, err
	}
	books, err := st.ListBooks()
	return books, scanErrs, err
}
```

Add `models` import:

```go
"github.com/xuanchong/cli-read/models"
```

- [ ] **Step 5: Use helpers in main startup**

Replace single-dir scan flow with:

```go
defaultCreated, err := ensureDefaultLibraryDir(st, paths.DefaultBookDir)
if err != nil {
	fmt.Fprintf(os.Stderr, "初始化默认书籍目录失败: %v\n", err)
	os.Exit(1)
}

scanDirs, err := configuredScanDirs(st, paths.TempBookDir)
if err != nil {
	fmt.Fprintf(os.Stderr, "读取书籍目录失败: %v\n", err)
	os.Exit(1)
}

books, scanErrs, err := refreshBooks(st, scanDirs)
if err != nil {
	fmt.Fprintf(os.Stderr, "同步书架失败: %v\n", err)
	os.Exit(1)
}
```

Set root fields:

```go
root := rootModel{
	dataDir:        paths.DataDir,
	defaultBookDir: paths.DefaultBookDir,
	tempBookDir:    paths.TempBookDir,
	store:          st,
	mode:           modeBookshelf,
	bookshelf:      app.NewBookshelfModel(books),
}
```

Add fields to `rootModel`:

```go
dataDir        string
defaultBookDir string
tempBookDir    string
```

Set status:

```go
status := fmt.Sprintf("已扫描 %d 本书", len(books))
if defaultCreated {
	status = fmt.Sprintf("未配置书籍目录，已使用默认目录 %s", paths.DefaultBookDir)
} else if len(scanErrs) > 0 {
	status = fmt.Sprintf("已扫描 %d 本书，%d 个目录扫描失败", len(books), len(scanErrs))
} else if paths.TempBookDir != "" {
	status = fmt.Sprintf("已临时扫描目录 %s；如需长期使用，请在目录管理中添加", paths.TempBookDir)
}
root.bookshelf.SetStatus(status)
```

- [ ] **Step 6: Run root tests**

Run:

```bash
go test . -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add main.go main_test.go
git commit -m "feat: 使用应用数据目录启动"
```

---

### Task 4: Directory Manager TUI Model

**Files:**
- Modify: `app/messages.go`
- Create: `app/directory_manager.go`
- Create: `app/directory_manager_test.go`

**Interfaces:**
- Consumes: `models.LibraryDir`, `config.NormalizePath`.
- Produces messages:
  - `type OpenDirectoryManagerMsg struct{}`
  - `type CloseDirectoryManagerMsg struct{}`
  - `type AddLibraryDirMsg struct { Path string; Create bool }`
  - `type DeleteLibraryDirMsg struct { Dir models.LibraryDir }`
  - `type RescanLibraryDirsMsg struct{}`

- [ ] **Step 1: Add failing directory manager tests**

Create `app/directory_manager_test.go`:

```go
package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xuanchong/cli-read/models"
)

func testDirs() []models.LibraryDir {
	return []models.LibraryDir{
		{ID: 1, Path: "/tmp/.cli-read/.book", IsDefault: true},
		{ID: 2, Path: "/tmp/books", IsDefault: false},
	}
}

func TestDirectoryManagerViewListsDirs(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	view := m.View()
	if !strings.Contains(view, "书籍目录") || !strings.Contains(view, "/tmp/books") || !strings.Contains(view, "默认") {
		t.Fatalf("view missing content: %s", view)
	}
}

func TestDirectoryManagerAddEmitsAddMessage(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(DirectoryManagerModel)
	if m.mode != dirModeAdd {
		t.Fatalf("mode: %d", m.mode)
	}
	for _, r := range "/tmp/new-books" {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(DirectoryManagerModel)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected add command")
	}
	msg := cmd()
	add, ok := msg.(AddLibraryDirMsg)
	if !ok {
		t.Fatalf("message: %T", msg)
	}
	if add.Path != "/tmp/new-books" || add.Create {
		t.Fatalf("add msg: %+v", add)
	}
}

func TestDirectoryManagerDeleteRequiresConfirmation(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(DirectoryManagerModel)
	if m.mode != dirModeConfirmDelete {
		t.Fatalf("mode: %d", m.mode)
	}
	view := m.View()
	if !strings.Contains(view, "阅读进度和书签") || !strings.Contains(view, "原始文件不会被删除") {
		t.Fatalf("confirm view: %s", view)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected delete command")
	}
	msg := cmd()
	del, ok := msg.(DeleteLibraryDirMsg)
	if !ok {
		t.Fatalf("message: %T", msg)
	}
	if del.Dir.ID != 1 {
		t.Fatalf("delete msg: %+v", del)
	}
}

func TestDirectoryManagerBackEmitsClose(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected close command")
	}
	if _, ok := cmd().(CloseDirectoryManagerMsg); !ok {
		t.Fatalf("message: %T", cmd())
	}
}

func TestDirectoryManagerRescanEmitsMessage(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected rescan command")
	}
	if _, ok := cmd().(RescanLibraryDirsMsg); !ok {
		t.Fatalf("message: %T", cmd())
	}
}
```

- [ ] **Step 2: Run failing app tests**

Run:

```bash
go test ./app -run DirectoryManager -v
```

Expected: FAIL because directory manager types are undefined.

- [ ] **Step 3: Add messages**

Append to `app/messages.go`:

```go
// OpenDirectoryManagerMsg is sent when the user opens book directory management.
type OpenDirectoryManagerMsg struct{}

// CloseDirectoryManagerMsg is sent when the user exits directory management.
type CloseDirectoryManagerMsg struct{}

// AddLibraryDirMsg asks the root model to add a managed book directory.
type AddLibraryDirMsg struct {
	Path   string
	Create bool
}

// DeleteLibraryDirMsg asks the root model to delete a managed directory and its book records.
type DeleteLibraryDirMsg struct {
	Dir models.LibraryDir
}

// RescanLibraryDirsMsg asks the root model to rescan all managed directories.
type RescanLibraryDirsMsg struct{}
```

- [ ] **Step 4: Implement directory manager model**

Create `app/directory_manager.go`:

```go
package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"

	"github.com/xuanchong/cli-read/models"
	"github.com/xuanchong/cli-read/ui"
)

type dirManagerMode int

const (
	dirModeList dirManagerMode = iota
	dirModeAdd
	dirModeConfirmDelete
)

type DirectoryManagerModel struct {
	dirs     []models.LibraryDir
	selected int
	mode     dirManagerMode
	input    textinput.Model
	keys     ui.KeyMap
	status   string
}

func NewDirectoryManagerModel(dirs []models.LibraryDir) DirectoryManagerModel {
	input := textinput.New()
	input.Placeholder = "粘贴或输入目录路径"
	input.CharLimit = 500
	return DirectoryManagerModel{
		dirs:  dirs,
		input: input,
		keys:  ui.DefaultKey(),
	}
}

func (m DirectoryManagerModel) Init() tea.Cmd { return nil }

func (m DirectoryManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case dirModeAdd:
			return m.updateAdd(msg)
		case dirModeConfirmDelete:
			return m.updateConfirmDelete(msg)
		default:
			return m.updateList(msg)
		}
	}
	return m, nil
}

func (m DirectoryManagerModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return CloseDirectoryManagerMsg{} }
	case key.Matches(msg, m.keys.Down):
		if m.selected+1 < len(m.dirs) {
			m.selected++
		}
	case key.Matches(msg, m.keys.Up):
		if m.selected > 0 {
			m.selected--
		}
	case msg.String() == "a":
		m.mode = dirModeAdd
		m.input.SetValue("")
		m.input.Focus()
	case key.Matches(msg, m.keys.Delete):
		if len(m.dirs) > 0 {
			m.mode = dirModeConfirmDelete
		}
	case msg.String() == "r":
		return m, func() tea.Msg { return RescanLibraryDirsMsg{} }
	}
	return m, nil
}

func (m DirectoryManagerModel) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = dirModeList
		m.input.Blur()
		return m, nil
	case key.Matches(msg, m.keys.Confirm):
		path := strings.TrimSpace(m.input.Value())
		m.mode = dirModeList
		m.input.Blur()
		if path == "" {
			m.status = "目录不能为空"
			return m, nil
		}
		return m, func() tea.Msg { return AddLibraryDirMsg{Path: path} }
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m DirectoryManagerModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if len(m.dirs) == 0 || m.selected >= len(m.dirs) {
			m.mode = dirModeList
			return m, nil
		}
		dir := m.dirs[m.selected]
		m.mode = dirModeList
		return m, func() tea.Msg { return DeleteLibraryDirMsg{Dir: dir} }
	case "esc", "q", "n":
		m.mode = dirModeList
		return m, nil
	}
	return m, nil
}

func (m DirectoryManagerModel) View() string {
	switch m.mode {
	case dirModeAdd:
		return "添加目录：\n" + m.input.View() + "\n\n" + ui.HintStyle.Render("Enter 保存  Esc/q 取消")
	case dirModeConfirmDelete:
		if len(m.dirs) == 0 || m.selected >= len(m.dirs) {
			return "没有可删除的目录\n\n" + ui.HintStyle.Render("Esc/q 返回")
		}
		dir := m.dirs[m.selected]
		return fmt.Sprintf("删除目录：%s\n\n这会删除该目录下已入库的书籍、阅读进度和书签。\n目录中的原始文件不会被删除。\n\n输入 y 确认删除，Esc/q 取消", dir.Path)
	default:
		var b strings.Builder
		b.WriteString("书籍目录\n\n")
		if len(m.dirs) == 0 {
			b.WriteString(ui.HintStyle.Render("（暂无目录）"))
			b.WriteString("\n")
		}
		for i, dir := range m.dirs {
			marker := "  "
			if i == m.selected {
				marker = "> "
			}
			b.WriteString(marker)
			b.WriteString(dir.Path)
			if dir.IsDefault {
				b.WriteString("        默认")
			}
			b.WriteString("\n")
		}
		if m.status != "" {
			b.WriteString("\n")
			b.WriteString(ui.StatusStyle.Render(m.status))
		}
		b.WriteString("\n")
		b.WriteString(ui.HintStyle.Render("a 添加目录  d 删除目录  r 重新扫描  q 返回书架"))
		 return b.String()
	}
}
```

- [ ] **Step 5: Fix formatting typo**

Run:

```bash
gofmt -w app/directory_manager.go app/directory_manager_test.go app/messages.go
```

Expected: no output. If gofmt fails because of the ` return b.String()` line, remove the extra leading space so the line is exactly:

```go
		return b.String()
```

Then rerun gofmt.

- [ ] **Step 6: Run app tests**

Run:

```bash
go test ./app -v
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add app/messages.go app/directory_manager.go app/directory_manager_test.go
git commit -m "feat: 添加目录管理界面"
```

---

### Task 5: Bookshelf Entry and Root Integration

**Files:**
- Modify: `ui/keys.go`
- Modify: `app/bookshelf.go`
- Modify: `app/bookshelf_test.go`
- Modify: `main.go`

**Interfaces:**
- Consumes messages from Task 4.
- Consumes store APIs from Task 2.
- Produces root handling for directory add/delete/rescan and bookshelf refresh.

- [ ] **Step 1: Add failing bookshelf shortcut test**

Append to `app/bookshelf_test.go`:

```go
func TestBookshelfOpenDirectoryManagerSendsMsg(t *testing.T) {
	books := []models.Book{{ID: 1, Title: "三体", Format: "epub"}}
	m := NewBookshelfModel(books)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("expected command")
	}
	if _, ok := cmd().(OpenDirectoryManagerMsg); !ok {
		t.Fatalf("expected OpenDirectoryManagerMsg")
	}
}
```

- [ ] **Step 2: Run failing bookshelf test**

Run:

```bash
go test ./app -run TestBookshelfOpenDirectoryManagerSendsMsg -v
```

Expected: FAIL because shortcut is not handled.

- [ ] **Step 3: Add key binding**

Modify `ui/keys.go`:

Add field:

```go
ManageDirs key.Binding
```

Add default binding:

```go
ManageDirs: key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "目录")),
```

- [ ] **Step 4: Handle bookshelf shortcut**

In `app/bookshelf.go`, inside the non-searching key switch, add before `Open`:

```go
case key.Matches(msg, m.keys.ManageDirs):
	return m, func() tea.Msg { return OpenDirectoryManagerMsg{} }
```

- [ ] **Step 5: Extend root model fields and modes**

In `main.go`, add mode:

```go
modeDirectoryManager
```

Add root field:

```go	directoryManager app.DirectoryManagerModel
```

- [ ] **Step 6: Add root refresh helpers**

Add to `main.go`:

```go
func (m rootModel) loadLibraryDirs() ([]models.LibraryDir, error) {
	return m.store.ListLibraryDirs()
}

func (m rootModel) refreshBookshelf(status string) (rootModel, error) {
	if _, err := ensureDefaultLibraryDir(m.store, m.defaultBookDir); err != nil {
		return m, err
	}
	dirs, err := configuredScanDirs(m.store, m.tempBookDir)
	if err != nil {
		return m, err
	}
	books, scanErrs, err := refreshBooks(m.store, dirs)
	if err != nil {
		return m, err
	}
	m.bookshelf = app.NewBookshelfModel(books)
	if status == "" {
		status = fmt.Sprintf("已扫描 %d 本书", len(books))
	}
	if len(scanErrs) > 0 {
		status = fmt.Sprintf("%s，%d 个目录扫描失败", status, len(scanErrs))
	}
	m.bookshelf.SetStatus(status)
	return m, nil
}

func (m rootModel) reloadDirectoryManager(status string) rootModel {
	dirs, err := m.store.ListLibraryDirs()
	if err != nil {
		m.directoryManager = app.NewDirectoryManagerModel(nil)
		return m
	}
	m.directoryManager = app.NewDirectoryManagerModel(dirs)
	return m
}
```

- [ ] **Step 7: Handle directory manager messages in root Update**

In `rootModel.Update`, add cases near `OpenBookMsg`:

```go
case app.OpenDirectoryManagerMsg:
	dirs, err := m.store.ListLibraryDirs()
	if err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("读取目录失败: %v", err))
		return m, nil
	}
	m.directoryManager = app.NewDirectoryManagerModel(dirs)
	m.mode = modeDirectoryManager
	return m, nil

case app.CloseDirectoryManagerMsg:
	m.mode = modeBookshelf
	return m, nil

case app.RescanLibraryDirsMsg:
	var err error
	m, err = m.refreshBookshelf("目录已重新扫描")
	if err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("扫描失败: %v", err))
	}
	m = m.reloadDirectoryManager("")
	m.mode = modeDirectoryManager
	return m, nil

case app.AddLibraryDirMsg:
	path, err := config.NormalizePath(msg.Path)
	if err != nil {
		m.directoryManager = app.NewDirectoryManagerModel(nil)
		return m, nil
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		if msg.Create {
			if err := os.MkdirAll(path, 0755); err != nil {
				m.bookshelf.SetStatus(fmt.Sprintf("创建目录失败: %v", err))
				return m, nil
			}
		} else {
			m.bookshelf.SetStatus(fmt.Sprintf("目录不存在: %s", path))
			return m, nil
		}
	} else if !info.IsDir() {
		m.bookshelf.SetStatus(fmt.Sprintf("不是目录: %s", path))
		return m, nil
	}
	if _, err := m.store.AddLibraryDir(path, false); err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("添加目录失败: %v", err))
		return m, nil
	}
	m, err = m.refreshBookshelf(fmt.Sprintf("已添加目录 %s", path))
	if err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("扫描失败: %v", err))
	}
	m = m.reloadDirectoryManager("")
	m.mode = modeDirectoryManager
	return m, nil

case app.DeleteLibraryDirMsg:
	if err := m.store.DeleteLibraryDir(msg.Dir.ID); err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("删除目录失败: %v", err))
		return m, nil
	}
	if err := m.store.DeleteBooksUnderDir(msg.Dir.Path); err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("删除书籍记录失败: %v", err))
		return m, nil
	}
	var err error
	m, err = m.refreshBookshelf(fmt.Sprintf("已删除目录 %s", msg.Dir.Path))
	if err != nil {
		m.bookshelf.SetStatus(fmt.Sprintf("扫描失败: %v", err))
	}
	m = m.reloadDirectoryManager("")
	m.mode = modeDirectoryManager
	return m, nil
```

- [ ] **Step 8: Route modeDirectoryManager messages**

In `rootModel.Update`, before bookshelf routing, add:

```go
if m.mode == modeDirectoryManager {
	dm, cmd := m.directoryManager.Update(msg)
	m.directoryManager = dm.(app.DirectoryManagerModel)
	return m, cmd
}
```

In `rootModel.View`, add:

```go
if m.mode == modeDirectoryManager {
	return m.directoryManager.View()
}
```

- [ ] **Step 9: Run app and root tests**

Run:

```bash
go test ./app -v
go test . -v
```

Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add ui/keys.go app/bookshelf.go app/bookshelf_test.go main.go
git commit -m "feat: 集成目录管理入口"
```

---

### Task 6: Documentation and End-to-End Verification

**Files:**
- Modify: `README.md`
- Modify: `CLAUDE.md`
- Modify: `e2e/e2e_test.go`

**Interfaces:**
- Consumes all tasks.
- Produces documented user behavior.

- [ ] **Step 1: Add e2e test for directory deletion cleanup**

Append to `e2e/e2e_test.go`:

```go
func TestManagedDirectoryDeleteRemovesBooksAndProgress(t *testing.T) {
	dir := t.TempDir()
	bookPath := filepath.Join(dir, "managed.txt")
	if err := writeFile(bookPath, []byte("Managed\n\n第一章\n内容")); err != nil {
		t.Fatalf("write book: %v", err)
	}

	st, err := store.Open(filepath.Join(t.TempDir(), "e2e.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()

	id, err := st.AddLibraryDir(dir, true)
	if err != nil {
		t.Fatalf("add dir: %v", err)
	}
	book, err := parser.ParseByExtension(bookPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	book.FilePath = bookPath
	bookID, err := st.UpsertBook(*book)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := st.SaveProgress(models.ReadingProgress{BookID: bookID, Chapter: 1, Page: 2}); err != nil {
		t.Fatalf("progress: %v", err)
	}
	if _, err := st.AddBookmark(models.Bookmark{BookID: bookID, Chapter: 0, Page: 0, Label: "mark"}); err != nil {
		t.Fatalf("bookmark: %v", err)
	}
	if err := st.DeleteLibraryDir(id); err != nil {
		t.Fatalf("delete dir: %v", err)
	}
	if err := st.DeleteBooksUnderDir(dir); err != nil {
		t.Fatalf("delete books: %v", err)
	}
	books, err := st.ListBooks()
	if err != nil {
		t.Fatalf("books: %v", err)
	}
	if len(books) != 0 {
		t.Fatalf("books remain: %+v", books)
	}
	if _, err := readFile(bookPath); err != nil {
		t.Fatalf("original file should remain: %v", err)
	}
}
```

- [ ] **Step 2: Update README usage and storage docs**

Update `README.md` usage section to include:

```markdown
# 默认使用 ~/.cli-read/novel-reader.db 和 ~/.cli-read/.book
reader

# 指定应用数据目录
reader --data-dir /path/to/app-data

# 临时扫描一个书籍目录（不保存到目录列表）
reader --dir /path/to/books

# 高级：指定数据库文件
reader --db /path/to/db.sqlite
```

Update data storage section to include:

```markdown
默认应用数据目录为 `~/.cli-read`：

```text
~/.cli-read/
├── novel-reader.db
└── .book/
```

书籍目录列表保存在 SQLite 的 `library_dirs` 表中。首次启动没有配置目录时，会自动使用 `~/.cli-read/.book`。
```

Update shortcuts:

```markdown
| `D` | 管理书籍目录 |
```

Add directory manager shortcut table:

```markdown
### 目录管理

| 键 | 功能 |
|----|------|
| `j` / `↓` | 向下移动 |
| `k` / `↑` | 向上移动 |
| `a` | 添加目录 |
| `d` | 删除目录及相关书籍记录 |
| `r` | 重新扫描 |
| `q` / `Esc` | 返回书架 |
```

- [ ] **Step 3: Update CLAUDE.md key design decisions**

Add bullets:

```markdown
- App data defaults to `~/.cli-read` (`novel-reader.db` + `.book`)
- Managed book directories are stored in SQLite `library_dirs`
- Directory deletion removes related book rows and cascades progress/bookmarks; original files are not deleted
```

- [ ] **Step 4: Run all tests**

Run:

```bash
go test ./... -v
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add README.md CLAUDE.md e2e/e2e_test.go
git commit -m "docs: 更新目录管理说明"
```

---

### Task 7: Final Verification

**Files:**
- All modified files

**Interfaces:**
- Consumes all prior tasks.
- Produces a verified working tree ready for final review.

- [ ] **Step 1: Format all Go files**

Run:

```bash
gofmt -w config/paths.go config/paths_test.go models/book.go store/library_dirs.go store/store_test.go main.go main_test.go app/messages.go app/directory_manager.go app/directory_manager_test.go app/bookshelf.go app/bookshelf_test.go ui/keys.go e2e/e2e_test.go
```

Expected: no output.

- [ ] **Step 2: Run full test suite**

Run:

```bash
go test ./... -v
```

Expected: PASS.

- [ ] **Step 3: Inspect final status**

Run:

```bash
git status --short
git log --oneline -5
```

Expected: working tree contains only intentional changes if commits were skipped by an implementer, or is clean if each task committed.

- [ ] **Step 4: Commit any uncommitted final formatting**

If `git status --short` shows modified files from formatting, run:

```bash
git add -A
git commit -m "chore: 格式化目录管理实现"
```

Expected: either a formatting commit is created or no commit is needed.

---

## Self-Review

- Spec coverage: default data directory, database path, default `.book`, `--data-dir`, `--db`, `--dir`, path normalization, SQLite `library_dirs`, startup default insertion, missing-dir scan tolerance, independent directory manager model, add/delete/rescan UI, destructive directory deletion, immediate refresh, docs, and tests are each covered by tasks.
- Placeholder scan: no TBD/TODO placeholders are present; every task includes exact paths, commands, and concrete code snippets.
- Type consistency: `models.LibraryDir`, store methods, app messages, and root helper names are consistent across tasks.
- Scope check: the plan is a single coherent feature with layered deliverables; each task can be reviewed independently.
