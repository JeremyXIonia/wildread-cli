# Task 2: Store 层（SQLite）

**Goal:** Create SQLite store layer with CRUD operations for books, reading_progress, and bookmarks tables.

**Pre-requisites:** Task 1 completed — `models/book.go` defines Book, Chapter, ReadingProgress, Bookmark structs.

**Output files:**
- `store/schema.sql`
- `store/store.go`
- `store/books.go`
- `store/progress.go`
- `store/bookmarks.go`
- `store/store_test.go`

## Global Constraints (relevant)
- Module: `github.com/xuanchong/cli-read`
- Use `github.com/mattn/go-sqlite3` (CGO required)
- Schema uses `ON DELETE CASCADE` for referential integrity
- Import the data models from `github.com/xuanchong/cli-read/models`

## Steps

### Step 1: Write schema.sql
```sql
CREATE TABLE IF NOT EXISTS books (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path   TEXT NOT NULL UNIQUE,
    title       TEXT NOT NULL,
    author      TEXT,
    format      TEXT NOT NULL,
    added_at    DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reading_progress (
    book_id     INTEGER PRIMARY KEY REFERENCES books(id) ON DELETE CASCADE,
    chapter     INTEGER DEFAULT 0,
    page        INTEGER DEFAULT 0,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id     INTEGER REFERENCES books(id) ON DELETE CASCADE,
    chapter     INTEGER NOT NULL,
    page        INTEGER NOT NULL,
    label       TEXT,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Step 2: Write store_test.go with all tests
```go
package store

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/xuanchong/cli-read/models"
)

func newTestStore(t *testing.T) *Store {
    t.Helper()
    path := filepath.Join(t.TempDir(), "test.db")
    s, err := Open(path)
    if err != nil { t.Fatalf("open: %v", err) }
    t.Cleanup(func() { s.Close(); os.Remove(path) })
    return s
}

func TestOpenClose(t *testing.T) {
    s := newTestStore(t)
    if err := s.Close(); err != nil { t.Fatalf("close: %v", err) }
}

func TestBooksCRUD(t *testing.T) {
    s := newTestStore(t)
    b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
    id, err := s.UpsertBook(b)
    if err != nil || id <= 0 { t.Fatalf("upsert: %v, %d", err, id) }
    got, _ := s.GetBook(id)
    if got.Title != "A" || got.Format != "epub" { t.Fatalf("bad book: %+v", got) }
    list, _ := s.ListBooks()
    if len(list) != 1 { t.Fatalf("list: %d", len(list)) }
    s.DeleteBook(id)
    list, _ = s.ListBooks()
    if len(list) != 0 { t.Fatalf("not empty") }
}

func TestUpsertBookUpdates(t *testing.T) {
    s := newTestStore(t)
    b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
    id, _ := s.UpsertBook(b)
    b.ID = id; b.Title = "A-new"
    id2, _ := s.UpsertBook(b)
    if id != id2 { t.Fatalf("id mismatch") }
    got, _ := s.GetBook(id)
    if got.Title != "A-new" { t.Fatalf("not updated") }
}

func TestProgressCRUD(t *testing.T) {
    s := newTestStore(t)
    id, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
    p, _ := s.GetProgress(id)
    if p.Chapter != 0 || p.Page != 0 { t.Fatalf("default: %+v", p) }
    s.SaveProgress(models.ReadingProgress{BookID: id, Chapter: 2, Page: 5})
    p, _ = s.GetProgress(id)
    if p.Chapter != 2 || p.Page != 5 { t.Fatalf("saved: %+v", p) }
}

func TestBookmarksCRUD(t *testing.T) {
    s := newTestStore(t)
    bid, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
    id1, _ := s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 1, Page: 2, Label: "m1"})
    s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 3, Page: 4, Label: "m2"})
    list, _ := s.ListBookmarks(bid)
    if len(list) != 2 || list[0].Label != "m1" { t.Fatalf("list: %+v", list) }
    s.DeleteBookmark(id1)
    list, _ = s.ListBookmarks(bid)
    if len(list) != 1 { t.Fatalf("after delete: %d", len(list)) }
}
```

### Step 3: Write store/store.go
```go
package store

import (
    "database/sql"
    _ "embed"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
    db *sql.DB
}

func Open(path string) (*Store, error) {
    db, err := sql.Open("sqlite3", path+"?_fk=1")
    if err != nil { return nil, fmt.Errorf("open db: %w", err) }
    if _, err := db.Exec(schemaSQL); err != nil {
        db.Close()
        return nil, fmt.Errorf("init schema: %w", err)
    }
    return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
```

### Step 4: Write store/books.go
```go
package store

import (
    "database/sql"
    "errors"
    "github.com/xuanchong/cli-read/models"
)

func (s *Store) UpsertBook(b models.Book) (int64, error) {
    if b.ID == 0 {
        res, err := s.db.Exec(
            `INSERT INTO books (file_path, title, author, format) VALUES (?, ?, ?, ?)`,
            b.FilePath, b.Title, b.Author, b.Format)
        if err != nil { return 0, err }
        return res.LastInsertId()
    }
    _, err := s.db.Exec(
        `UPDATE books SET title=?, author=?, format=? WHERE id=?`,
        b.Title, b.Author, b.Format, b.ID)
    return b.ID, err
}

func (s *Store) GetBook(id int64) (models.Book, error) {
    var b models.Book
    err := s.db.QueryRow(
        `SELECT id, file_path, title, COALESCE(author, ''), format FROM books WHERE id=?`,
        id).Scan(&b.ID, &b.FilePath, &b.Title, &b.Author, &b.Format)
    if errors.Is(err, sql.ErrNoRows) {
        return b, errors.New("not found")
    }
    return b, err
}

func (s *Store) ListBooks() ([]models.Book, error) {
    rows, err := s.db.Query(
        `SELECT id, file_path, title, COALESCE(author, ''), format FROM books ORDER BY added_at`)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []models.Book
    for rows.Next() {
        var b models.Book
        if err := rows.Scan(&b.ID, &b.FilePath, &b.Title, &b.Author, &b.Format); err != nil {
            return nil, err
        }
        out = append(out, b)
    }
    return out, rows.Err()
}

func (s *Store) DeleteBook(id int64) error {
    _, err := s.db.Exec(`DELETE FROM books WHERE id=?`, id)
    return err
}
```

### Step 5: Write store/progress.go
```go
package store

import (
    "database/sql"
    "errors"
    "github.com/xuanchong/cli-read/models"
)

func (s *Store) GetProgress(bookID int64) (models.ReadingProgress, error) {
    var p models.ReadingProgress
    p.BookID = bookID
    err := s.db.QueryRow(
        `SELECT chapter, page FROM reading_progress WHERE book_id=?`,
        bookID).Scan(&p.Chapter, &p.Page)
    if errors.Is(err, sql.ErrNoRows) { return p, nil }
    return p, err
}

func (s *Store) SaveProgress(p models.ReadingProgress) error {
    _, err := s.db.Exec(`
        INSERT INTO reading_progress (book_id, chapter, page, updated_at)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(book_id) DO UPDATE SET
            chapter=excluded.chapter, page=excluded.page, updated_at=CURRENT_TIMESTAMP`,
        p.BookID, p.Chapter, p.Page)
    return err
}
```

### Step 6: Write store/bookmarks.go
```go
package store

import "github.com/xuanchong/cli-read/models"

func (s *Store) AddBookmark(b models.Bookmark) (int64, error) {
    res, err := s.db.Exec(
        `INSERT INTO bookmarks (book_id, chapter, page, label) VALUES (?, ?, ?, ?)`,
        b.BookID, b.Chapter, b.Page, b.Label)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (s *Store) ListBookmarks(bookID int64) ([]models.Bookmark, error) {
    rows, err := s.db.Query(
        `SELECT id, book_id, chapter, page, COALESCE(label, ''), created_at
         FROM bookmarks WHERE book_id=? ORDER BY id`, bookID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []models.Bookmark
    for rows.Next() {
        var b models.Bookmark
        if err := rows.Scan(&b.ID, &b.BookID, &b.Chapter, &b.Page, &b.Label, &b.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, b)
    }
    return out, rows.Err()
}

func (s *Store) DeleteBookmark(id int64) error {
    _, err := s.db.Exec(`DELETE FROM bookmarks WHERE id=?`, id)
    return err
}
```

### Step 7: Run tests
`go test ./store/... -v` — should PASS all 5 tests

### Step 8: Commit
`git add store/ && git commit -m "feat: store 层（SQLite + books/progress/bookmarks CRUD）"`

## Report
Write report to `docs/superpowers/plans/task-2-report.md` with status, commits, test results, concerns.
