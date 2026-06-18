# 终端小说阅读器实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建一个跨 Windows 和 macOS 的终端小说阅读器，支持 EPUB/TXT/Markdown 格式，提供书架、章节目录、分页阅读、阅读进度记录、书签等核心功能。

**Architecture:** 分层架构 — `parser/` 解析文档为统一 `Book/Chapter` 模型，`store/` 用 SQLite 持久化状态，`pager/` 负责文本分页，`app/` 用 bubbletea 实现 TUI（书架、阅读、目录、书签四个界面），`main.go` 串联所有组件并提供命令行参数。

**Tech Stack:** Go 1.22+、Bubble Tea (TUI)、Bubble Components、go-sqlite3、golang.org/x/text（编码）、golang.org/x/net/html（HTML 解析）、goldmark（Markdown 解析）

## Global Constraints

- Go 版本：≥ 1.22
- 跨平台：Windows 10+ 和 macOS 12+
- 数据存储：单个 SQLite 文件，默认 `./novel-reader.db`
- 书籍目录：CLI 参数 `--dir` 指定，默认 `./books`
- 界面语言：中文
- 配色：跟随终端，不自定义颜色
- 键盘风格：Vim 风格（j/k/gg/G/o/q 等）
- 格式支持：EPUB、TXT、Markdown 三种
- 编码支持：UTF-8、GBK、GB18030（自动检测）
- Markdown 渲染：提取纯文本，保留段落（用空行分隔）
- EPUB 解析：自实现（archive/zip + golang.org/x/net/html）
- 书架扫描：启动时单次扫描，不做实时监控
- 范围之外：自定义配色、实时文件监控、云同步、复杂 Markdown 渲染、表格代码块
- 测试框架：标准库 `testing`
- 提交粒度：每完成一个任务做一次 `git commit`

## File Structure

```
cli-read/
├── main.go                       # 入口：CLI 参数 + 主循环
├── go.mod
├── go.sum
├── README.md
├── docs/
│   └── superpowers/
│       ├── specs/2026-06-18-terminal-novel-reader-design.md
│       └── plans/2026-06-18-terminal-novel-reader.md
├── app/
│   ├── bookshelf.go              # 书架 UI
│   ├── reader.go                 # 阅读器 UI
│   ├── messages.go
│   ├── bookshelf_test.go
│   └── reader_test.go
├── models/
│   ├── book.go
│   └── book_test.go
├── store/
│   ├── store.go
│   ├── schema.sql
│   ├── books.go
│   ├── progress.go
│   ├── bookmarks.go
│   └── store_test.go
├── parser/
│   ├── parser.go
│   ├── epub.go
│   ├── txt.go
│   ├── markdown.go
│   ├── html2text.go
│   ├── scanner.go
│   ├── iohelper.go
│   └── *_test.go
├── pager/
│   ├── pager.go
│   └── pager_test.go
├── ui/
│   ├── keys.go
│   └── styles.go
├── e2e/
│   ├── e2e_test.go
│   └── io.go
├── testdata/
│   ├── sample.txt
│   ├── sample.md
│   ├── sample.epub
│   └── make_sample_epub.sh
└── build.sh
```

---

## Task 1: 项目骨架与数据模型

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `models/book.go`
- Create: `models/book_test.go`

**Interfaces:**
- Produces: `models.Book`, `models.Chapter`, `models.ReadingProgress`, `models.Bookmark`

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd D:\workspace-latest\cli-read
go mod init github.com/xuanchong/cli-read
```

- [ ] **Step 2: 添加依赖**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/mattn/go-sqlite3@latest
go get golang.org/x/text@latest
go get golang.org/x/net/html@latest
go get github.com/yuin/goldmark@latest
```

- [ ] **Step 3: 编写数据模型 `models/book.go`**

```go
package models

type Book struct {
    ID       int64
    FilePath string
    Title    string
    Author   string
    Format   string
}

type Chapter struct {
    Title   string
    Content string
}

type ReadingProgress struct {
    BookID  int64
    Chapter int
    Page    int
}

type Bookmark struct {
    ID        int64
    BookID    int64
    Chapter   int
    Page      int
    Label     string
    CreatedAt string
}
```

- [ ] **Step 4: 编写测试 `models/book_test.go`**

```go
package models

import "testing"

func TestBookFields(t *testing.T) {
    b := Book{ID: 1, FilePath: "/x.epub", Title: "X", Format: "epub"}
    if b.ID != 1 || b.Title != "X" || b.Format != "epub" {
        t.Fatalf("unexpected: %+v", b)
    }
}

func TestChapterContent(t *testing.T) {
    c := Chapter{Title: "ch1", Content: "p1\n\np2"}
    if c.Title != "ch1" || c.Content != "p1\n\np2" {
        t.Fatalf("unexpected: %+v", c)
    }
}
```

- [ ] **Step 5: 运行测试**

```bash
go test ./models/...
```

Expected: PASS

- [ ] **Step 6: 编写 main.go 骨架**

```go
package main

import (
    "flag"
    "fmt"
    "os"
)

func main() {
    dir := flag.String("dir", "./books", "书籍目录")
    dbPath := flag.String("db", "./novel-reader.db", "SQLite 数据库路径")
    flag.Parse()

    fmt.Printf("书籍目录: %s\n", *dir)
    fmt.Printf("数据库: %s\n", *dbPath)

    if _, err := os.Stat(*dir); os.IsNotExist(err) {
        fmt.Fprintf(os.Stderr, "目录不存在: %s\n", *dir)
        os.Exit(1)
    }
}
```

- [ ] **Step 7: 编译并运行**

```bash
go build -o reader.exe .
./reader.exe --dir ./books
```

- [ ] **Step 8: 提交**

```bash
git init
git add go.mod go.sum main.go models/
git commit -m "feat: 项目骨架与数据模型"
```

---

## Task 2: Store 层（SQLite）

**Files:**
- Create: `store/schema.sql`
- Create: `store/store.go`
- Create: `store/books.go`
- Create: `store/progress.go`
- Create: `store/bookmarks.go`
- Create: `store/store_test.go`

**Interfaces:**
- `store.Open(path) (*Store, error)`
- `(*Store).Close() error`
- `(*Store).UpsertBook(b) (int64, error)`
- `(*Store).GetBook(id) (models.Book, error)`
- `(*Store).ListBooks() ([]models.Book, error)`
- `(*Store).DeleteBook(id) error`
- `(*Store).GetProgress(bookID) (models.ReadingProgress, error)`
- `(*Store).SaveProgress(p) error`
- `(*Store).AddBookmark(b) (int64, error)`
- `(*Store).ListBookmarks(bookID) ([]models.Bookmark, error)`
- `(*Store).DeleteBookmark(id) error`

- [ ] **Step 1: 编写 `store/schema.sql`**

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

- [ ] **Step 2: 编写 `store/store_test.go`（含全部测试）**

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
    if err != nil {
        t.Fatalf("open: %v", err)
    }
    t.Cleanup(func() {
        s.Close()
        os.Remove(path)
    })
    return s
}

func TestOpenClose(t *testing.T) {
    s := newTestStore(t)
    if err := s.Close(); err != nil {
        t.Fatalf("close: %v", err)
    }
}

func TestBooksCRUD(t *testing.T) {
    s := newTestStore(t)
    b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
    id, err := s.UpsertBook(b)
    if err != nil || id <= 0 {
        t.Fatalf("upsert: %v, %d", err, id)
    }
    got, _ := s.GetBook(id)
    if got.Title != "A" || got.Format != "epub" {
        t.Fatalf("bad book: %+v", got)
    }
    list, _ := s.ListBooks()
    if len(list) != 1 {
        t.Fatalf("list len: %d", len(list))
    }
    s.DeleteBook(id)
    list, _ = s.ListBooks()
    if len(list) != 0 {
        t.Fatalf("expected empty")
    }
}

func TestUpsertBookUpdates(t *testing.T) {
    s := newTestStore(t)
    b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
    id, _ := s.UpsertBook(b)
    b.ID = id
    b.Title = "A-new"
    id2, _ := s.UpsertBook(b)
    if id != id2 {
        t.Fatalf("id mismatch")
    }
    got, _ := s.GetBook(id)
    if got.Title != "A-new" {
        t.Fatalf("not updated: %+v", got)
    }
}

func TestProgressCRUD(t *testing.T) {
    s := newTestStore(t)
    id, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
    p, _ := s.GetProgress(id)
    if p.Chapter != 0 || p.Page != 0 {
        t.Fatalf("default: %+v", p)
    }
    s.SaveProgress(models.ReadingProgress{BookID: id, Chapter: 2, Page: 5})
    p, _ = s.GetProgress(id)
    if p.Chapter != 2 || p.Page != 5 {
        t.Fatalf("saved: %+v", p)
    }
}

func TestBookmarksCRUD(t *testing.T) {
    s := newTestStore(t)
    bid, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
    id1, _ := s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 1, Page: 2, Label: "m1"})
    s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 3, Page: 4, Label: "m2"})
    list, _ := s.ListBookmarks(bid)
    if len(list) != 2 || list[0].Label != "m1" {
        t.Fatalf("list: %+v", list)
    }
    s.DeleteBookmark(id1)
    list, _ = s.ListBookmarks(bid)
    if len(list) != 1 {
        t.Fatalf("after delete: %d", len(list))
    }
}
```

- [ ] **Step 3: 运行测试，应失败**

```bash
go test ./store/...
```

Expected: FAIL（Open 未定义）

- [ ] **Step 4: 实现 `store/store.go`**

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
    if err != nil {
        return nil, fmt.Errorf("open db: %w", err)
    }
    if _, err := db.Exec(schemaSQL); err != nil {
        db.Close()
        return nil, fmt.Errorf("init schema: %w", err)
    }
    return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }
```

- [ ] **Step 5: 实现 `store/books.go`**

```go
package store

import (
    "database/sql"
    "errors"
    "fmt"

    "github.com/xuanchong/cli-read/models"
)

func (s *Store) UpsertBook(b models.Book) (int64, error) {
    if b.ID == 0 {
        res, err := s.db.Exec(
            `INSERT INTO books (file_path, title, author, format) VALUES (?, ?, ?, ?)`,
            b.FilePath, b.Title, b.Author, b.Format,
        )
        if err != nil {
            return 0, err
        }
        return res.LastInsertId()
    }
    _, err := s.db.Exec(
        `UPDATE books SET title=?, author=?, format=? WHERE id=?`,
        b.Title, b.Author, b.Format, b.ID,
    )
    return b.ID, err
}

func (s *Store) GetBook(id int64) (models.Book, error) {
    var b models.Book
    err := s.db.QueryRow(
        `SELECT id, file_path, title, COALESCE(author, ''), format FROM books WHERE id=?`,
        id,
    ).Scan(&b.ID, &b.FilePath, &b.Title, &b.Author, &b.Format)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return b, fmt.Errorf("book %d not found", id)
        }
        return b, err
    }
    return b, nil
}

func (s *Store) ListBooks() ([]models.Book, error) {
    rows, err := s.db.Query(
        `SELECT id, file_path, title, COALESCE(author, ''), format FROM books ORDER BY added_at`,
    )
    if err != nil {
        return nil, err
    }
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

- [ ] **Step 6: 实现 `store/progress.go`**

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
        bookID,
    ).Scan(&p.Chapter, &p.Page)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return p, nil
        }
        return p, err
    }
    return p, nil
}

func (s *Store) SaveProgress(p models.ReadingProgress) error {
    _, err := s.db.Exec(`
        INSERT INTO reading_progress (book_id, chapter, page, updated_at)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(book_id) DO UPDATE SET
            chapter=excluded.chapter,
            page=excluded.page,
            updated_at=CURRENT_TIMESTAMP
    `, p.BookID, p.Chapter, p.Page)
    return err
}
```

- [ ] **Step 7: 实现 `store/bookmarks.go`**

```go
package store

import "github.com/xuanchong/cli-read/models"

func (s *Store) AddBookmark(b models.Bookmark) (int64, error) {
    res, err := s.db.Exec(
        `INSERT INTO bookmarks (book_id, chapter, page, label) VALUES (?, ?, ?, ?)`,
        b.BookID, b.Chapter, b.Page, b.Label,
    )
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}

func (s *Store) ListBookmarks(bookID int64) ([]models.Bookmark, error) {
    rows, err := s.db.Query(
        `SELECT id, book_id, chapter, page, COALESCE(label, ''), created_at
         FROM bookmarks WHERE book_id=? ORDER BY id`,
        bookID,
    )
    if err != nil {
        return nil, err
    }
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

- [ ] **Step 8: 运行所有 store 测试**

```bash
go test ./store/... -v
```

Expected: PASS

- [ ] **Step 9: 提交**

```bash
git add store/
git commit -m "feat: store 层（SQLite + books/progress/bookmarks CRUD）"
```

---

## Task 3: Pager 层（文本分页）

**Files:**
- Create: `pager/pager.go`
- Create: `pager/pager_test.go`

**Interfaces:**
- `pager.New(text, width, height) *Pager`
- `(*Pager).PageCount() int`
- `(*Pager).Page(idx) (string, error)`
- `(*Pager).LineWidth() int`

- [ ] **Step 1: 编写 `pager/pager_test.go`**

```go
package pager

import "testing"

func TestPagerEnglish(t *testing.T) {
    text := "line one..\nline two..\nline three\nline four.\nline five."
    p := New(text, 10, 3)
    if got := p.PageCount(); got != 2 {
        t.Fatalf("page count: %d", got)
    }
    p1, _ := p.Page(0)
    want1 := "line one..\nline two..\nline three"
    if p1 != want1 {
        t.Fatalf("page 0: %q", p1)
    }
    p2, _ := p.Page(1)
    want2 := "line four.\nline five."
    if p2 != want2 {
        t.Fatalf("page 1: %q", p2)
    }
}

func TestPagerChinese(t *testing.T) {
    text := "你好世界\n你好世界\n你好世界"
    p := New(text, 8, 2)
    if got := p.PageCount(); got != 2 {
        t.Fatalf("page count: %d", got)
    }
}

func TestPagerEmpty(t *testing.T) {
    p := New("", 10, 5)
    if p.PageCount() != 1 {
        t.Fatalf("empty count: %d", p.PageCount())
    }
    c, _ := p.Page(0)
    if c != "" {
        t.Fatalf("empty: %q", c)
    }
}

func TestPagerOutOfRange(t *testing.T) {
    p := New("a\nb\nc", 10, 2)
    if _, err := p.Page(-1); err == nil {
        t.Fatal("expected error for -1")
    }
    if _, err := p.Page(99); err == nil {
        t.Fatal("expected error for 99")
    }
}
```

- [ ] **Step 2: 运行测试，应失败**

```bash
go test ./pager/...
```

Expected: FAIL

- [ ] **Step 3: 实现 `pager/pager.go`**

```go
package pager

import (
    "fmt"
    "strings"
)

type Pager struct {
    lines    []string
    width    int
    pageSize int
}

func runeWidth(r rune) int {
    if r < 0x80 {
        return 1
    }
    return 2
}

func displayWidth(s string) int {
    w := 0
    for _, r := range s {
        w += runeWidth(r)
    }
    return w
}

func wrap(line string, width int) []string {
    if displayWidth(line) <= width {
        return []string{line}
    }
    var out []string
    var cur []rune
    curW := 0
    for _, r := range []rune(line) {
        rw := runeWidth(r)
        if curW+rw > width && len(cur) > 0 {
            out = append(out, string(cur))
            cur = cur[:0]
            curW = 0
        }
        cur = append(cur, r)
        curW += rw
    }
    if len(cur) > 0 {
        out = append(out, string(cur))
    }
    return out
}

func New(text string, width, height int) *Pager {
    if width <= 0 {
        width = 80
    }
    if height <= 0 {
        height = 24
    }
    raw := strings.Split(text, "\n")
    var lines []string
    for _, l := range raw {
        l = strings.TrimRight(l, "\r")
        if l == "" {
            lines = append(lines, "")
            continue
        }
        lines = append(lines, wrap(l, width)...)
    }
    return &Pager{lines: lines, width: width, pageSize: height}
}

func (p *Pager) PageCount() int {
    if len(p.lines) == 0 {
        return 1
    }
    n := len(p.lines) / p.pageSize
    if len(p.lines)%p.pageSize != 0 {
        n++
    }
    if n < 1 {
        n = 1
    }
    return n
}

func (p *Pager) Page(idx int) (string, error) {
    n := p.PageCount()
    if idx < 0 || idx >= n {
        return "", fmt.Errorf("page %d out of range [0, %d)", idx, n)
    }
    start := idx * p.pageSize
    end := start + p.pageSize
    if end > len(p.lines) {
        end = len(p.lines)
    }
    return strings.Join(p.lines[start:end], "\n"), nil
}

func (p *Pager) LineWidth() int { return p.width }
```

- [ ] **Step 4: 运行测试**

```bash
go test ./pager/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add pager/
git commit -m "feat: pager 层（文本分页，支持中文）"
```

---

## Task 4: Parser 公共接口

**Files:**
- Create: `parser/parser.go`
- Create: `parser/parser_test.go`
- Create: `parser/epub.go`（占位）
- Create: `parser/txt.go`（占位）
- Create: `parser/markdown.go`（占位）

**Interfaces:**
- `parser.ParseByExtension(path) (*models.Book, error)`
- `parser.FormatFromExt(path) string`

- [ ] **Step 1: 创建占位文件**

`parser/epub.go`:
```go
package parser

import "github.com/xuanchong/cli-read/models"

func ParseEPUB(path string) (*models.Book, error) {
    return &models.Book{Format: "epub"}, nil
}
```

`parser/txt.go`:
```go
package parser

import "github.com/xuanchong/cli-read/models"

func ParseTXT(path string) (*models.Book, error) {
    return &models.Book{Format: "txt"}, nil
}
```

`parser/markdown.go`:
```go
package parser

import "github.com/xuanchong/cli-read/models"

func ParseMarkdown(path string) (*models.Book, error) {
    return &models.Book{Format: "md"}, nil
}
```

- [ ] **Step 2: 编写 `parser/parser.go`**

```go
package parser

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/xuanchong/cli-read/models"
)

func FormatFromExt(path string) string {
    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".epub":
        return "epub"
    case ".txt":
        return "txt"
    case ".md", ".markdown":
        return "md"
    default:
        return ""
    }
}

func ParseByExtension(path string) (*models.Book, error) {
    switch FormatFromExt(path) {
    case "epub":
        return ParseEPUB(path)
    case "txt":
        return ParseTXT(path)
    case "md":
        return ParseMarkdown(path)
    default:
        return nil, fmt.Errorf("unsupported format: %s", path)
    }
}
```

- [ ] **Step 3: 编写 `parser/parser_test.go`**

```go
package parser

import "testing"

func TestFormatFromExt(t *testing.T) {
    cases := map[string]string{
        "a.epub":     "epub",
        "a.txt":      "txt",
        "a.md":       "md",
        "a.markdown": "md",
        "a.zip":      "",
        "a":          "",
    }
    for in, want := range cases {
        if got := FormatFromExt(in); got != want {
            t.Errorf("%s: got %q, want %q", in, got, want)
        }
    }
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./parser/...
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add parser/
git commit -m "feat: parser 公共接口 + 扩展名分派"
```

---

## Task 5: HTML→纯文本工具

**Files:**
- Create: `parser/html2text.go`
- Create: `parser/html2text_test.go`

**Interfaces:**
- `parser.HTMLToText(html) (title string, paragraphs []string)`

- [ ] **Step 1: 编写 `parser/html2text_test.go`**

```go
package parser

import "testing"

func TestHTMLToTextParagraphs(t *testing.T) {
    html := `<html><body><p>第一段。</p><p>第二段。</p></body></html>`
    title, paras := HTMLToText(html)
    if title != "" {
        t.Errorf("title: %q", title)
    }
    if len(paras) != 2 || paras[0] != "第一段。" || paras[1] != "第二段。" {
        t.Fatalf("paragraphs: %+v", paras)
    }
}

func TestHTMLToTextHeading(t *testing.T) {
    html := `<html><body><h1>第一章</h1><p>正文。</p></body></html>`
    title, paras := HTMLToText(html)
    if title != "第一章" {
        t.Errorf("title: %q", title)
    }
    if len(paras) != 1 || paras[0] != "正文。" {
        t.Fatalf("paragraphs: %+v", paras)
    }
}

func TestHTMLToTextStripTags(t *testing.T) {
    html := `<p>这是 <b>加粗</b> 和 <i>斜体</i> 的文本。</p>`
    _, paras := HTMLToText(html)
    if len(paras) != 1 || paras[0] != "这是 加粗 和 斜体 的文本。" {
        t.Fatalf("paragraphs: %+v", paras)
    }
}

func TestHTMLToTextEmpty(t *testing.T) {
    title, paras := HTMLToText("")
    if title != "" || len(paras) != 0 {
        t.Fatalf("empty: %q %v", title, paras)
    }
}
```

- [ ] **Step 2: 运行测试，应失败**

```bash
go test ./parser/...
```

Expected: FAIL

- [ ] **Step 3: 实现 `parser/html2text.go`**

```go
package parser

import (
    "strings"

    "golang.org/x/net/html"
)

func HTMLToText(input string) (string, []string) {
    doc, err := html.Parse(strings.NewReader(input))
    if err != nil {
        return "", nil
    }

    var title string
    var paragraphs []string
    var current strings.Builder

    var walk func(n *html.Node)
    walk = func(n *html.Node) {
        if n.Type == html.TextNode {
            current.WriteString(n.Data)
            return
        }
        if n.Type != html.ElementNode {
            return
        }
        tag := strings.ToLower(n.Data)
        if tag == "script" || tag == "style" {
            return
        }
        if tag == "br" {
            current.WriteString("\n")
        }
        for c := n.FirstChild; c != nil; c = c.NextSibling {
            walk(c)
        }
        isHeading := tag == "h1" || tag == "h2" || tag == "h3" || tag == "h4" || tag == "h5" || tag == "h6"
        if isHeading && title == "" {
            t := strings.TrimSpace(current.String())
            if t != "" {
                title = t
                current.Reset()
                return
            }
        }
        if tag == "p" || tag == "div" {
            text := strings.TrimSpace(current.String())
            if text != "" {
                paragraphs = append(paragraphs, text)
            }
            current.Reset()
        }
    }

    walk(doc)
    if t := strings.TrimSpace(current.String()); t != "" && title == "" {
        title = t
    }
    return title, paragraphs
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./parser/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add parser/html2text.go parser/html2text_test.go
git commit -m "feat: HTML→纯文本工具"
```

---

## Task 6: TXT 解析器

**Files:**
- Modify: `parser/txt.go`（覆盖占位）
- Create: `parser/iohelper.go`
- Create: `parser/txt_test.go`
- Create: `testdata/sample.txt`

- [ ] **Step 1: 创建 `testdata/sample.txt`（UTF-8 无 BOM）**

```
第一段内容。

第二段内容。

第三段内容。
```

- [ ] **Step 2: 编写 `parser/txt_test.go`**

```go
package parser

import (
    "path/filepath"
    "strings"
    "testing"
)

func TestParseTXTBasic(t *testing.T) {
    book, err := ParseTXT(filepath.Join("..", "testdata", "sample.txt"))
    if err != nil {
        t.Fatalf("parse: %v", err)
    }
    if book.Title == "" {
        t.Fatal("title empty")
    }
    if len(book.Chapters) != 1 {
        t.Fatalf("chapters: %d", len(book.Chapters))
    }
    c := book.Chapters[0]
    if c.Title != "" {
        t.Errorf("chapter title: %q", c.Title)
    }
    if !strings.Contains(c.Content, "第一段内容。") ||
        !strings.Contains(c.Content, "第二段内容。") ||
        !strings.Contains(c.Content, "第三段内容。") {
        t.Fatalf("content: %q", c.Content)
    }
}
```

- [ ] **Step 3: 运行测试，应失败**

```bash
go test ./parser/... -run TestParseTXT
```

Expected: FAIL

- [ ] **Step 4: 实现 `parser/txt.go`**

```go
package parser

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "unicode/utf8"

    "github.com/xuanchong/cli-read/models"
    "golang.org/x/text/transform"
)

func ParseTXT(path string) (*models.Book, error) {
    raw, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read txt: %w", err)
    }
    raw = stripBOM(raw)
    text, err := decodeBytes(raw)
    if err != nil {
        return nil, fmt.Errorf("decode txt: %w", err)
    }
    text = strings.ReplaceAll(text, "\r\n", "\n")
    text = strings.ReplaceAll(text, "\r", "\n")
    blocks := strings.Split(text, "\n\n")
    var paragraphs []string
    for _, b := range blocks {
        b = strings.TrimSpace(b)
        if b != "" {
            paragraphs = append(paragraphs, b)
        }
    }
    content := strings.Join(paragraphs, "\n\n")
    base := filepath.Base(path)
    title := strings.TrimSuffix(base, filepath.Ext(base))
    return &models.Book{
        Title:    title,
        Author:   "",
        Format:   "txt",
        Chapters: []models.Chapter{{Title: "", Content: content}},
    }, nil
}

func stripBOM(b []byte) []byte {
    if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
        return b[3:]
    }
    return b
}

func decodeBytes(b []byte) (string, error) {
    if utf8.Valid(b) {
        return string(b), nil
    }
    for _, label := range []string{"gb18030", "gbk"} {
        enc, err := lookupEncoding(label)
        if err != nil {
            continue
        }
        reader := transform.NewReader(newBytesReader(b), enc.NewDecoder())
        decoded, err := readAll(reader)
        if err == nil {
            return decoded, nil
        }
    }
    return "", fmt.Errorf("unable to detect encoding")
}
```

- [ ] **Step 5: 创建 `parser/iohelper.go`**

```go
package parser

import (
    "bytes"
    "io"

    "golang.org/x/text/encoding"
    "golang.org/x/text/encoding/htmlindex"
)

func newBytesReader(b []byte) io.Reader { return bytes.NewReader(b) }

func readAll(r io.Reader) (string, error) {
    b, err := io.ReadAll(r)
    return string(b), err
}

func lookupEncoding(label string) (encoding.Encoding, error) {
    return htmlindex.Get(label)
}
```

- [ ] **Step 6: 运行测试**

```bash
go test ./parser/... -v
```

Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add parser/txt.go parser/iohelper.go parser/txt_test.go testdata/sample.txt
git commit -m "feat: TXT 解析器（编码自动检测）"
```

---

## Task 7: Markdown 解析器

**Files:**
- Modify: `parser/markdown.go`（覆盖占位）
- Create: `parser/markdown_test.go`
- Create: `testdata/sample.md`

- [ ] **Step 1: 创建 `testdata/sample.md`**

```markdown
# 第一章 开始

这是第一章的第一段。

这是第一章的第二段。

# 第二章 继续

这是第二章的内容。
```

- [ ] **Step 2: 编写 `parser/markdown_test.go`**

```go
package parser

import (
    "path/filepath"
    "strings"
    "testing"
)

func TestParseMarkdown(t *testing.T) {
    book, err := ParseMarkdown(filepath.Join("..", "testdata", "sample.md"))
    if err != nil {
        t.Fatalf("parse: %v", err)
    }
    if book.Title == "" {
        t.Fatal("title empty")
    }
    if len(book.Chapters) != 2 {
        t.Fatalf("chapters: %d", len(book.Chapters))
    }
    if book.Chapters[0].Title != "第一章 开始" {
        t.Errorf("ch0 title: %q", book.Chapters[0].Title)
    }
    if !strings.Contains(book.Chapters[0].Content, "第一章的第一段") {
        t.Errorf("ch0 content: %q", book.Chapters[0].Content)
    }
    if book.Chapters[1].Title != "第二章 继续" {
        t.Errorf("ch1 title: %q", book.Chapters[1].Title)
    }
}
```

- [ ] **Step 3: 运行测试，应失败**

```bash
go test ./parser/... -run TestParseMarkdown
```

Expected: FAIL

- [ ] **Step 4: 实现 `parser/markdown.go`**

```go
package parser

import (
    "bytes"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/xuanchong/cli-read/models"
    "github.com/yuin/goldmark"
    "github.com/yuin/goldmark/ast"
    "github.com/yuin/goldmark/text"
)

var md = goldmark.New()

func ParseMarkdown(path string) (*models.Book, error) {
    raw, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read md: %w", err)
    }
    reader := text.NewReader(raw)
    root := md.Parser().Parse(reader)

    var chapters []models.Chapter
    var current models.Chapter
    flush := func() {
        if current.Title != "" || strings.TrimSpace(current.Content) != "" {
            current.Content = strings.TrimSpace(current.Content)
            chapters = append(chapters, current)
        }
        current = models.Chapter{}
    }

    err = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
        if !entering {
            return ast.WalkContinue, nil
        }
        switch v := n.(type) {
        case *ast.Heading:
            flush()
            level := v.Level
            txt := nodeText(v, raw)
            if level <= 2 {
                current.Title = strings.TrimSpace(txt)
            } else {
                current.Content += strings.Repeat("#", level) + " " + txt + "\n\n"
            }
        case *ast.Paragraph:
            current.Content += strings.TrimSpace(nodeText(v, raw)) + "\n\n"
        }
        return ast.WalkContinue, nil
    })
    if err != nil {
        return nil, err
    }
    flush()

    base := filepath.Base(path)
    title := strings.TrimSuffix(base, filepath.Ext(base))
    return &models.Book{
        Title:    title,
        Format:   "md",
        Chapters: chapters,
    }, nil
}

func nodeText(n ast.Node, source []byte) string {
    var buf bytes.Buffer
    for i := 0; i < n.Lines().Len(); i++ {
        seg := n.Lines().At(i)
        buf.Write(seg.Value(source))
    }
    return buf.String()
}
```

- [ ] **Step 5: 运行测试**

```bash
go test ./parser/... -v
```

Expected: PASS

- [ ] **Step 6: 提交**

```bash
git add parser/markdown.go parser/markdown_test.go testdata/sample.md
git commit -m "feat: Markdown 解析器（按 #/## 切章节）"
```

---

## Task 8: EPUB 解析器

**Files:**
- Modify: `parser/epub.go`（覆盖占位）
- Create: `parser/epub_test.go`
- Create: `testdata/make_sample_epub.sh`
- Create: `testdata/sample.epub`（由脚本生成）

- [ ] **Step 1: 创建 `testdata/make_sample_epub.sh`**

```bash
#!/bin/bash
set -e
DIR=$(mktemp -d)
trap "rm -rf $DIR" EXIT

mkdir -p "$DIR/META-INF" "$DIR/OEBPS"

cat > "$DIR/META-INF/container.xml" <<'EOF'
<?xml version="1.0"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
EOF

cat > "$DIR/OEBPS/content.opf" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="bid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="bid">urn:uuid:00000000-0000-0000-0000-000000000001</dc:identifier>
    <dc:title>测试书</dc:title>
    <dc:creator>测试作者</dc:creator>
    <dc:language>zh</dc:language>
  </metadata>
  <manifest>
    <item id="ch1" href="ch1.xhtml" media-type="application/xhtml+xml"/>
    <item id="ch2" href="ch2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="ch1"/>
    <itemref idref="ch2"/>
  </spine>
</package>
EOF

cat > "$DIR/OEBPS/ch1.xhtml" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>第一章</title></head>
<body><h1>第一章 开始</h1><p>第一段内容。</p><p>第二段内容。</p></body></html>
EOF

cat > "$DIR/OEBPS/ch2.xhtml" <<'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml"><head><title>第二章</title></head>
<body><h1>第二章 继续</h1><p>这是第二章的内容。</p></body></html>
EOF

cd "$DIR"
zip -r "$OLDPWD/testdata/sample.epub" META-INF OEBPS
echo "Created sample.epub"
```

- [ ] **Step 2: 生成 sample.epub**

```bash
chmod +x testdata/make_sample_epub.sh
bash testdata/make_sample_epub.sh
```

Expected: `testdata/sample.epub` 存在

- [ ] **Step 3: 编写 `parser/epub_test.go`**

```go
package parser

import (
    "path/filepath"
    "strings"
    "testing"
)

func TestParseEPUB(t *testing.T) {
    book, err := ParseEPUB(filepath.Join("..", "testdata", "sample.epub"))
    if err != nil {
        t.Fatalf("parse: %v", err)
    }
    if book.Title != "测试书" {
        t.Errorf("title: %q", book.Title)
    }
    if book.Author != "测试作者" {
        t.Errorf("author: %q", book.Author)
    }
    if len(book.Chapters) != 2 {
        t.Fatalf("chapters: %d", len(book.Chapters))
    }
    if book.Chapters[0].Title != "第一章 开始" {
        t.Errorf("ch0 title: %q", book.Chapters[0].Title)
    }
    if !strings.Contains(book.Chapters[0].Content, "第一段内容") {
        t.Errorf("ch0 content: %q", book.Chapters[0].Content)
    }
    if book.Chapters[1].Title != "第二章 继续" {
        t.Errorf("ch1 title: %q", book.Chapters[1].Title)
    }
}
```

- [ ] **Step 4: 运行测试，应失败**

```bash
go test ./parser/... -run TestParseEPUB
```

Expected: FAIL

- [ ] **Step 5: 实现 `parser/epub.go`**

```go
package parser

import (
    "archive/zip"
    "bytes"
    "encoding/xml"
    "fmt"
    "io"
    "path/filepath"
    "strings"

    "github.com/xuanchong/cli-read/models"
)

func ParseEPUB(path string) (*models.Book, error) {
    zr, err := zip.OpenReader(path)
    if err != nil {
        return nil, fmt.Errorf("open epub: %w", err)
    }
    defer zr.Close()

    var opfPath string
    for _, f := range zr.File {
        if f.Name == "META-INF/container.xml" {
            data, err := readZipFile(f)
            if err != nil {
                return nil, err
            }
            var cont struct {
                Rootfiles struct {
                    Rootfile struct {
                        FullPath string `xml:"full-path,attr"`
                    } `xml:"rootfile"`
                } `xml:"rootfiles"`
            }
            if err := xml.Unmarshal(data, &cont); err != nil {
                return nil, fmt.Errorf("parse container: %w", err)
            }
            opfPath = cont.Rootfiles.Rootfile.FullPath
            break
        }
    }
    if opfPath == "" {
        return nil, fmt.Errorf("opf not found in container")
    }

    var opfData []byte
    for _, f := range zr.File {
        if f.Name == opfPath {
            data, err := readZipFile(f)
            if err != nil {
                return nil, err
            }
            opfData = data
            break
        }
    }
    if opfData == nil {
        return nil, fmt.Errorf("opf not found: %s", opfPath)
    }

    type opfPackage struct {
        Metadata struct {
            Title   string `xml:"title"`
            Creator string `xml:"creator"`
        } `xml:"metadata"`
        Manifest struct {
            Items []struct {
                ID        string `xml:"id,attr"`
                Href      string `xml:"href,attr"`
                MediaType string `xml:"media-type,attr"`
            } `xml:"item"`
        } `xml:"manifest"`
        Spine struct {
            Items []struct {
                IDRef string `xml:"idref,attr"`
            } `xml:"itemref"`
        } `xml:"spine"`
    }
    var pkg opfPackage
    if err := xml.Unmarshal(opfData, &pkg); err != nil {
        return nil, fmt.Errorf("parse opf: %w", err)
    }

    opfDir := filepath.Dir(opfPath)
    idToHref := map[string]string{}
    for _, it := range pkg.Manifest.Items {
        idToHref[it.ID] = it.Href
    }

    var chapters []models.Chapter
    for _, ref := range pkg.Spine.Items {
        href, ok := idToHref[ref.IDRef]
        if !ok {
            continue
        }
        lower := strings.ToLower(href)
        if !strings.Contains(lower, ".xhtml") && !strings.Contains(lower, ".html") {
            continue
        }
        fullPath := filepath.ToSlash(filepath.Join(opfDir, href))
        var fileData []byte
        for _, f := range zr.File {
            if f.Name == fullPath {
                d, err := readZipFile(f)
                if err != nil {
                    return nil, err
                }
                fileData = d
                break
            }
        }
        if fileData == nil {
            continue
        }
        title, paras := HTMLToText(string(fileData))
        content := strings.Join(paras, "\n\n")
        chapters = append(chapters, models.Chapter{Title: title, Content: content})
    }

    return &models.Book{
        Title:    pkg.Metadata.Title,
        Author:   pkg.Metadata.Creator,
        Format:   "epub",
        Chapters: chapters,
    }, nil
}

func readZipFile(f *zip.File) ([]byte, error) {
    rc, err := f.Open()
    if err != nil {
        return nil, err
    }
    defer rc.Close()
    var buf bytes.Buffer
    if _, err := io.Copy(&buf, rc); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}
```

- [ ] **Step 6: 运行测试**

```bash
go test ./parser/... -v
```

Expected: PASS

- [ ] **Step 7: 提交**

```bash
git add parser/epub.go parser/epub_test.go testdata/make_sample_epub.sh testdata/sample.epub
git commit -m "feat: EPUB 解析器（自实现 ZIP + OPF 解析）"
```

---

## Task 9: 目录扫描器

**Files:**
- Create: `parser/scanner.go`
- Create: `parser/scanner_test.go`

**Interfaces:**
- `parser.Scan(dir) ([]string, error)`

- [ ] **Step 1: 编写 `parser/scanner_test.go`**

```go
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
        t.Fatalf("expected 3, got %d: %v", len(got), got)
    }
    found := map[string]bool{}
    for _, p := range got {
        found[FormatFromExt(p)] = true
    }
    if !found["epub"] || !found["txt"] || !found["md"] {
        t.Fatalf("missing: %v", found)
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
```

- [ ] **Step 2: 运行测试，应失败**

```bash
go test ./parser/... -run TestScan
```

Expected: FAIL

- [ ] **Step 3: 实现 `parser/scanner.go`**

```go
package parser

import (
    "fmt"
    "os"
    "path/filepath"
)

func Scan(dir string) ([]string, error) {
    info, err := os.Stat(dir)
    if err != nil {
        return nil, fmt.Errorf("stat dir: %w", err)
    }
    if !info.IsDir() {
        return nil, fmt.Errorf("%s is not a directory", dir)
    }
    var out []string
    err = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        if FormatFromExt(path) != "" {
            out = append(out, path)
        }
        return nil
    })
    if err != nil {
        return nil, err
    }
    return out, nil
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./parser/... -v
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add parser/scanner.go parser/scanner_test.go
git commit -m "feat: 目录扫描器"
```

---

## Task 10: 键盘定义和样式

**Files:**
- Create: `ui/keys.go`
- Create: `ui/styles.go`

- [ ] **Step 1: 编写 `ui/keys.go`**

```go
package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
    Up        key.Binding
    Down      key.Binding
    PageUp    key.Binding
    PageDown  key.Binding
    GotoTop   key.Binding
    GotoEnd   key.Binding
    Open      key.Binding
    Search    key.Binding
    Back      key.Binding
    Quit      key.Binding
    Next      key.Binding
    Prev      key.Binding
    Mark      key.Binding
    Bookmarks key.Binding
    Confirm   key.Binding
    Delete    key.Binding
}

func DefaultKey() KeyMap {
    return KeyMap{
        Up:        key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "上移")),
        Down:      key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "下移")),
        PageUp:    key.NewBinding(key.WithKeys("pgup", "b"), key.WithHelp("pgup", "上页")),
        PageDown:  key.NewBinding(key.WithKeys("pgdown", "f", " "), key.WithHelp("space", "下页")),
        GotoTop:   key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "到首")),
        GotoEnd:   key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "到尾")),
        Open:      key.NewBinding(key.WithKeys("o", "enter"), key.WithHelp("o/enter", "打开")),
        Search:    key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "搜索")),
        Back:      key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc/q", "返回")),
        Quit:      key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "退出")),
        Next:      key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "下一章")),
        Prev:      key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "上一章")),
        Mark:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "加书签")),
        Bookmarks: key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "书签")),
        Confirm:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "确认")),
        Delete:    key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "删除")),
    }
}
```

- [ ] **Step 2: 编写 `ui/styles.go`**

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
    TitleStyle    = lipgloss.NewStyle().Bold(true).Padding(0, 1)
    StatusStyle   = lipgloss.NewStyle().Faint(true)
    SelectedStyle = lipgloss.NewStyle().Bold(true)
    HintStyle     = lipgloss.NewStyle().Faint(true)
)
```

- [ ] **Step 3: 编译验证**

```bash
go build ./...
```

Expected: 无错误

- [ ] **Step 4: 提交**

```bash
git add ui/
git commit -m "feat: 全局键盘映射和样式"
```

---

## Task 11: 书架 UI

**Files:**
- Create: `app/messages.go`
- Create: `app/bookshelf.go`
- Create: `app/bookshelf_test.go`

**Interfaces:**
- `app.NewBookshelfModel(books) BookshelfModel`
- `(*BookshelfModel).Update(msg) (tea.Model, tea.Cmd)`
- `(*BookshelfModel).View() string`
- `(*BookshelfModel).Selected() *models.Book`
- `(*BookshelfModel).SetStatus(s)`

- [ ] **Step 1: 编写 `app/messages.go`**

```go
package app

import "github.com/xuanchong/cli-read/models"

type OpenBookMsg struct {
    Book models.Book
}
```

- [ ] **Step 2: 编写 `app/bookshelf.go`**

```go
package app

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/textinput"

    "github.com/xuanchong/cli-read/models"
    "github.com/xuanchong/cli-read/ui"
)

type bookItem struct {
    book models.Book
}

func (b bookItem) Title() string       { return b.book.Title }
func (b bookItem) Description() string { return fmt.Sprintf("[%s] %s", strings.ToUpper(b.book.Format), b.book.Author) }
func (b bookItem) FilterValue() string { return b.book.Title }

type BookshelfModel struct {
    list      list.Model
    input     textinput.Model
    keys      ui.KeyMap
    searching bool
    status    string
    allItems  []list.Item
}

func NewBookshelfModel(books []models.Book) BookshelfModel {
    items := make([]list.Item, len(books))
    for i, b := range books {
        items[i] = bookItem{book: b}
    }
    l := list.New(items, list.NewDefaultDelegate(), 60, 20)
    l.Title = "书架"
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(false)
    ti := textinput.New()
    ti.Placeholder = "搜索书名..."
    ti.CharLimit = 50
    return BookshelfModel{
        list:     l,
        input:    ti,
        keys:     ui.DefaultKey(),
        allItems: items,
    }
}

func (m *BookshelfModel) SetStatus(s string) { m.status = s }

func (m BookshelfModel) Init() tea.Cmd { return nil }

func (m BookshelfModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.searching {
        switch msg := msg.(type) {
        case tea.KeyMsg:
            switch msg.String() {
            case "esc":
                m.searching = false
                m.input.Blur()
                m.input.SetValue("")
            case "enter":
                term := m.input.Value()
                m.searching = false
                m.input.Blur()
                m.input.SetValue("")
                m.filter(term)
            default:
                var cmd tea.Cmd
                m.input, cmd = m.input.Update(msg)
                return m, cmd
            }
        }
        return m, nil
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.list.SetSize(msg.Width, msg.Height-4)
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m, tea.Quit
        case key.Matches(msg, m.keys.Search):
            m.searching = true
            m.input.Focus()
        case key.Matches(msg, m.keys.Open):
            if item, ok := m.list.SelectedItem().(bookItem); ok {
                return m, func() tea.Msg { return OpenBookMsg{Book: item.book} }
            }
        }
    }

    var cmd tea.Cmd
    m.list, cmd = m.list.Update(msg)
    return m, cmd
}

func (m *BookshelfModel) filter(term string) {
    term = strings.ToLower(term)
    if term == "" {
        m.list.SetItems(m.allItems)
        return
    }
    var items []list.Item
    for _, it := range m.allItems {
        bi := it.(bookItem)
        if strings.Contains(strings.ToLower(bi.book.Title), term) {
            items = append(items, bi)
        }
    }
    m.list.SetItems(items)
}

func (m BookshelfModel) View() string {
    var b strings.Builder
    b.WriteString(m.list.View())
    b.WriteString("\n")
    if m.searching {
        b.WriteString("/ ")
        b.WriteString(m.input.View())
    } else if m.status != "" {
        b.WriteString(ui.StatusStyle.Render(m.status))
    }
    return b.String()
}

func (m BookshelfModel) Selected() *models.Book {
    if item, ok := m.list.SelectedItem().(bookItem); ok {
        b := item.book
        return &b
    }
    return nil
}
```

- [ ] **Step 3: 编写 `app/bookshelf_test.go`**

```go
package app

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/xuanchong/cli-read/models"
)

func TestBookshelfModelInit(t *testing.T) {
    books := []models.Book{
        {ID: 1, Title: "三体", Format: "epub"},
        {ID: 2, Title: "活着", Format: "txt"},
    }
    m := NewBookshelfModel(books)
    got := m.Selected()
    if got == nil || got.Title != "三体" {
        t.Fatalf("expected 三体, got %+v", got)
    }
}

func TestBookshelfOpenSendsMsg(t *testing.T) {
    books := []models.Book{{ID: 1, Title: "三体", Format: "epub"}}
    m := NewBookshelfModel(books)
    _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
    if cmd == nil {
        t.Fatal("expected command")
    }
    msg := cmd()
    openMsg, ok := msg.(OpenBookMsg)
    if !ok {
        t.Fatalf("expected OpenBookMsg, got %T", msg)
    }
    if openMsg.Book.Title != "三体" {
        t.Errorf("book: %+v", openMsg.Book)
    }
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./app/...
```

Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add app/messages.go app/bookshelf.go app/bookshelf_test.go
git commit -m "feat: 书架 UI（Vim 风格导航 + 搜索）"
```

---

## Task 12: 阅读器 UI

**Files:**
- Create: `app/reader.go`
- Create: `app/reader_test.go`

**Interfaces:**
- `app.NewReaderModel(book, progress, store) ReaderModel`
- `(*ReaderModel).Update(msg) (tea.Model, tea.Cmd)`
- `(*ReaderModel).View() string`

- [ ] **Step 1: 编写 `app/reader.go`**

```go
package app

import (
    "fmt"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/viewport"

    "github.com/xuanchong/cli-read/models"
    "github.com/xuanchong/cli-read/pager"
    "github.com/xuanchong/cli-read/store"
    "github.com/xuanchong/cli-read/ui"
)

type ReaderMode int

const (
    ModeReading ReaderMode = iota
    ModeTOC
    ModeBookmarks
)

type ReaderModel struct {
    book      *models.Book
    pager     *pager.Pager
    viewport  viewport.Model
    keys      ui.KeyMap
    chapter   int
    page      int
    width     int
    height    int
    mode      ReaderMode
    status    string
    store     *store.Store
    bookmarks []models.Bookmark
}

const (
    readerHeaderLines = 1
    readerFooterLines = 2
)

func NewReaderModel(book *models.Book, progress models.ReadingProgress, st *store.Store) ReaderModel {
    chapter := progress.Chapter
    page := progress.Page
    if chapter >= len(book.Chapters) {
        chapter = 0
    }
    p := pager.New(book.Chapters[chapter].Content, 80, 20)
    if page >= p.PageCount() {
        page = 0
    }
    vp := viewport.New(80, 18)
    var bms []models.Bookmark
    if st != nil {
        bms, _ = st.ListBookmarks(book.ID)
    }
    return ReaderModel{
        book:      book,
        pager:     p,
        viewport:  vp,
        keys:      ui.DefaultKey(),
        chapter:   chapter,
        page:      page,
        mode:      ModeReading,
        store:     st,
        bookmarks: bms,
    }
}

func (m ReaderModel) Init() tea.Cmd {
    m.loadPageContent()
    return m.saveProgress()
}

func (m *ReaderModel) loadPageContent() {
    content, err := m.pager.Page(m.page)
    if err != nil {
        m.status = "页码错误"
        return
    }
    m.viewport.SetContent(content)
    m.status = ""
}

func (m ReaderModel) saveProgress() tea.Cmd {
    return func() tea.Msg {
        if m.store == nil || m.book == nil {
            return nil
        }
        _ = m.store.SaveProgress(models.ReadingProgress{
            BookID:  m.book.ID,
            Chapter: m.chapter,
            Page:    m.page,
        })
        return nil
    }
}

func (m ReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        m.resize()
        return m, nil
    case tea.KeyMsg:
        switch m.mode {
        case ModeReading:
            return m.updateReading(msg)
        case ModeTOC:
            return m.updateTOC(msg)
        case ModeBookmarks:
            return m.updateBookmarks(msg)
        }
    }
    var cmd tea.Cmd
    m.viewport, cmd = m.viewport.Update(msg)
    return m, cmd
}

func (m ReaderModel) updateReading(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch {
    case key.Matches(msg, m.keys.Back):
        return m, m.saveProgress()
    case key.Matches(msg, m.keys.PageDown):
        m.nextPage()
    case key.Matches(msg, m.keys.PageUp):
        m.prevPage()
    case key.Matches(msg, m.keys.GotoTop):
        m.page = 0
        m.loadPageContent()
    case key.Matches(msg, m.keys.GotoEnd):
        m.page = m.pager.PageCount() - 1
        m.loadPageContent()
    case key.Matches(msg, m.keys.Next):
        m.nextChapter()
    case key.Matches(msg, m.keys.Prev):
        m.prevChapter()
    case key.Matches(msg, m.keys.Open):
        m.mode = ModeTOC
    case key.Matches(msg, m.keys.Bookmarks):
        if m.store != nil {
            m.bookmarks, _ = m.store.ListBookmarks(m.book.ID)
        }
        m.mode = ModeBookmarks
    case key.Matches(msg, m.keys.Mark):
        if m.store == nil {
            m.status = "无数据库"
            break
        }
        id, err := m.store.AddBookmark(models.Bookmark{
            BookID:  m.book.ID,
            Chapter: m.chapter,
            Page:    m.page,
            Label:   fmt.Sprintf("ch%d p%d", m.chapter+1, m.page+1),
        })
        if err != nil {
            m.status = "加书签失败"
        } else {
            m.status = fmt.Sprintf("已加书签 #%d", id)
        }
    }
    return m, m.saveProgress()
}

func (m ReaderModel) updateTOC(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if key.Matches(msg, m.keys.Back) {
        m.mode = ModeReading
    }
    return m, nil
}

func (m ReaderModel) updateBookmarks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    if key.Matches(msg, m.keys.Back) {
        m.mode = ModeReading
    }
    return m, nil
}

func (m *ReaderModel) nextPage() {
    if m.page+1 < m.pager.PageCount() {
        m.page++
    } else if m.chapter+1 < len(m.book.Chapters) {
        m.chapter++
        m.page = 0
        m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
    }
    m.loadPageContent()
}

func (m *ReaderModel) prevPage() {
    if m.page > 0 {
        m.page--
    } else if m.chapter > 0 {
        m.chapter--
        m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
        m.page = m.pager.PageCount() - 1
    }
    m.loadPageContent()
}

func (m *ReaderModel) nextChapter() {
    if m.chapter+1 < len(m.book.Chapters) {
        m.chapter++
        m.page = 0
        m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
    }
    m.loadPageContent()
}

func (m *ReaderModel) prevChapter() {
    if m.chapter > 0 {
        m.chapter--
        m.page = 0
        m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
    }
    m.loadPageContent()
}

func (m *ReaderModel) resize() {
    bodyHeight := m.height - readerHeaderLines - readerFooterLines
    if bodyHeight < 5 {
        bodyHeight = 5
    }
    bodyWidth := m.width
    if bodyWidth < 20 {
        bodyWidth = 20
    }
    m.viewport.Width = bodyWidth
    m.viewport.Height = bodyHeight
    m.pager = pager.New(m.book.Chapters[m.chapter].Content, bodyWidth, bodyHeight)
    if m.page >= m.pager.PageCount() {
        m.page = m.pager.PageCount() - 1
        if m.page < 0 {
            m.page = 0
        }
    }
    m.loadPageContent()
}

func (m ReaderModel) header() string {
    chTitle := m.book.Chapters[m.chapter].Title
    if chTitle == "" {
        chTitle = fmt.Sprintf("第 %d 章", m.chapter+1)
    }
    return fmt.Sprintf("《%s》 - %s", m.book.Title, chTitle)
}

func (m ReaderModel) footer() string {
    total := m.pager.PageCount()
    var b strings.Builder
    b.WriteString(fmt.Sprintf("第 %d/%d 页 | 第 %d/%d 章", m.page+1, total, m.chapter+1, len(m.book.Chapters)))
    if m.status != "" {
        b.WriteString("  ")
        b.WriteString(ui.StatusStyle.Render(m.status))
    }
    b.WriteString("  ")
    b.WriteString(ui.HintStyle.Render("j/k 翻页 o 目录 n 下一章 p 上一章 m 加书签 b 书签 q 返回"))
    return b.String()
}

func (m ReaderModel) View() string {
    if m.mode == ModeTOC {
        return m.viewTOC()
    }
    if m.mode == ModeBookmarks {
        return m.viewBookmarks()
    }
    var b strings.Builder
    b.WriteString(ui.TitleStyle.Render(m.header()))
    b.WriteString("\n")
    b.WriteString(m.viewport.View())
    b.WriteString("\n")
    b.WriteString(m.footer())
    return b.String()
}

func (m ReaderModel) viewTOC() string {
    var b strings.Builder
    b.WriteString("章节目录\n\n")
    for i, c := range m.book.Chapters {
        marker := "  "
        if i == m.chapter {
            marker = "> "
        }
        title := c.Title
        if title == "" {
            title = fmt.Sprintf("第 %d 章", i+1)
        }
        b.WriteString(fmt.Sprintf("%s%s\n", marker, title))
    }
    b.WriteString("\n")
    b.WriteString(ui.HintStyle.Render("esc/q 返回阅读"))
    return b.String()
}

func (m ReaderModel) viewBookmarks() string {
    var b strings.Builder
    b.WriteString("书签\n\n")
    if len(m.bookmarks) == 0 {
        b.WriteString(ui.HintStyle.Render("（暂无书签）"))
    } else {
        for _, bm := range m.bookmarks {
            b.WriteString(fmt.Sprintf("#%d  第 %d 章 第 %d 页  %s\n", bm.ID, bm.Chapter+1, bm.Page+1, bm.Label))
        }
    }
    b.WriteString("\n")
    b.WriteString(ui.HintStyle.Render("esc/q 返回阅读"))
    return b.String()
}
```

- [ ] **Step 2: 编写 `app/reader_test.go`**

```go
package app

import (
    "path/filepath"
    "testing"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/xuanchong/cli-read/models"
    "github.com/xuanchong/cli-read/parser"
    "github.com/xuanchong/cli-read/store"
)

func newTestReader(t *testing.T) (ReaderModel, *store.Store) {
    t.Helper()
    book, err := parser.ParseMarkdown(filepath.Join("..", "testdata", "sample.md"))
    if err != nil {
        t.Fatalf("parse md: %v", err)
    }
    st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
    if err != nil {
        t.Fatalf("open store: %v", err)
    }
    t.Cleanup(func() { st.Close() })

    bid, _ := st.UpsertBook(models.Book{FilePath: "/x.md", Title: book.Title, Format: "md"})
    book.ID = bid
    return NewReaderModel(book, models.ReadingProgress{BookID: bid}, st), st
}

func TestReaderInit(t *testing.T) {
    m, _ := newTestReader(t)
    if m.chapter != 0 || m.page != 0 {
        t.Fatalf("init: ch=%d page=%d", m.chapter, m.page)
    }
}

func TestReaderTOCMode(t *testing.T) {
    m, _ := newTestReader(t)
    updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
    rm := updated.(ReaderModel)
    if rm.mode != ModeTOC {
        t.Fatalf("expected TOC mode, got %d", rm.mode)
    }
    view := rm.View()
    if view == "" {
        t.Fatal("view empty")
    }
    if !contains(view, "章节目录") {
        t.Fatal("view missing TOC title")
    }
}

func contains(s, sub string) bool {
    if len(sub) == 0 {
        return true
    }
    for i := 0; i+len(sub) <= len(s); i++ {
        if s[i:i+len(sub)] == sub {
            return true
        }
    }
    return false
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./app/...
```

Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add app/reader.go app/reader_test.go
git commit -m "feat: 阅读器 UI（分页、目录、书签基础）"
```

---

## Task 13: 主程序集成

**Files:**
- Modify: `main.go`

- [ ] **Step 1: 重写 `main.go`**

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "path/filepath"

    tea "github.com/charmbracelet/bubbletea"

    "github.com/xuanchong/cli-read/app"
    "github.com/xuanchong/cli-read/parser"
    "github.com/xuanchong/cli-read/store"
)

type rootModel struct {
    dir       string
    store     *store.Store
    mode      appMode
    bookshelf app.BookshelfModel
    reader    *app.ReaderModel
}

type appMode int

const (
    modeBookshelf appMode = iota
    modeReader
)

func main() {
    dir := flag.String("dir", "./books", "书籍目录")
    dbPath := flag.String("db", "./novel-reader.db", "SQLite 数据库路径")
    flag.Parse()

    if err := os.MkdirAll(*dir, 0755); err != nil {
        fmt.Fprintf(os.Stderr, "无法创建书籍目录: %v\n", err)
        os.Exit(1)
    }
    st, err := store.Open(*dbPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "无法打开数据库: %v\n", err)
        os.Exit(1)
    }
    defer st.Close()

    paths, err := parser.Scan(*dir)
    if err != nil {
        fmt.Fprintf(os.Stderr, "扫描目录失败: %v\n", err)
        os.Exit(1)
    }
    if err := syncBooks(st, paths); err != nil {
        fmt.Fprintf(os.Stderr, "同步书架失败: %v\n", err)
        os.Exit(1)
    }
    books, err := st.ListBooks()
    if err != nil {
        fmt.Fprintf(os.Stderr, "读取书架失败: %v\n", err)
        os.Exit(1)
    }

    root := rootModel{
        dir:       *dir,
        store:     st,
        mode:      modeBookshelf,
        bookshelf: app.NewBookshelfModel(books),
    }
    root.bookshelf.SetStatus(fmt.Sprintf("已扫描 %d 本书", len(books)))

    p := tea.NewProgram(root, tea.WithAltScreen())
    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "运行错误: %v\n", err)
        os.Exit(1)
    }
}

func syncBooks(st *store.Store, paths []string) error {
    existing, err := st.ListBooks()
    if err != nil {
        return err
    }
    existingByPath := map[string]int64{}
    for _, b := range existing {
        existingByPath[b.FilePath] = b.ID
    }
    seen := map[string]bool{}
    for _, p := range paths {
        abs, err := filepath.Abs(p)
        if err != nil {
            abs = p
        }
        seen[abs] = true
        if _, ok := existingByPath[abs]; ok {
            continue
        }
        book, err := parser.ParseByExtension(abs)
        if err != nil {
            fmt.Fprintf(os.Stderr, "解析失败 %s: %v\n", abs, err)
            continue
        }
        book.FilePath = abs
        if _, err := st.UpsertBook(*book); err != nil {
            fmt.Fprintf(os.Stderr, "入库失败 %s: %v\n", abs, err)
        }
    }
    for _, b := range existing {
        if !seen[b.FilePath] {
            if err := st.DeleteBook(b.ID); err != nil {
                fmt.Fprintf(os.Stderr, "删除失败 %d: %v\n", b.ID, err)
            }
        }
    }
    return nil
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case app.OpenBookMsg:
        progress, _ := m.store.GetProgress(msg.Book.ID)
        reader := app.NewReaderModel(&msg.Book, progress, m.store)
        m.reader = &reader
        m.mode = modeReader
        return m, nil
    case tea.WindowSizeMsg:
        if m.mode == modeBookshelf {
            bs, cmd := m.bookshelf.Update(msg)
            m.bookshelf = bs.(app.BookshelfModel)
            return m, cmd
        }
    case tea.KeyMsg:
        if m.mode == modeReader && (msg.String() == "esc" || msg.String() == "q") {
            m.mode = modeBookshelf
            return m, nil
        }
    }
    if m.mode == modeBookshelf {
        bs, cmd := m.bookshelf.Update(msg)
        m.bookshelf = bs.(app.BookshelfModel)
        return m, cmd
    }
    if m.mode == modeReader && m.reader != nil {
        nm, cmd := m.reader.Update(msg)
        m.reader = &nm.(app.ReaderModel)
        return m, cmd
    }
    return m, nil
}

func (m rootModel) View() string {
    if m.mode == modeReader && m.reader != nil {
        return m.reader.View()
    }
    return m.bookshelf.View()
}
```

- [ ] **Step 2: 编译**

```bash
go build -o reader.exe .
```

Expected: 无错误，生成 `reader.exe`

- [ ] **Step 3: 准备测试书籍**

```bash
mkdir -p ./books
cp testdata/sample.md ./books/
cp testdata/sample.txt ./books/
cp testdata/sample.epub ./books/ 2>/dev/null || echo "EPUB not built"
```

- [ ] **Step 4: 手动验证**

```bash
./reader.exe --dir ./books
```

测试交互：`j`/`k` 移动、`/` 搜索、`Enter` 打开、`Space` 翻页、`o` 目录、`n`/`p` 章节、`m` 加书签、`b` 书签、`q` 返回/退出

- [ ] **Step 5: 提交**

```bash
git add main.go
git commit -m "feat: 主程序集成（CLI + 扫描 + 同步 + TUI 切换）"
```

---

## Task 14: 端到端测试

**Files:**
- Create: `e2e/e2e_test.go`
- Create: `e2e/io.go`

- [ ] **Step 1: 编写 `e2e/io.go`**

```go
package e2e

import "os"

func readFile(p string) ([]byte, error)  { return os.ReadFile(p) }
func writeFile(p string, d []byte) error { return os.WriteFile(p, d, 0644) }
```

- [ ] **Step 2: 编写 `e2e/e2e_test.go`**

```go
package e2e

import (
    "path/filepath"
    "testing"

    "github.com/xuanchong/cli-read/models"
    "github.com/xuanchong/cli-read/parser"
    "github.com/xuanchong/cli-read/store"
)

func TestE2EFlow(t *testing.T) {
    dir := t.TempDir()
    sources := []string{
        filepath.Join("..", "testdata", "sample.txt"),
        filepath.Join("..", "testdata", "sample.md"),
        filepath.Join("..", "testdata", "sample.epub"),
    }
    var paths []string
    for _, src := range sources {
        if _, err := readFile(src); err != nil {
            t.Skipf("missing source %s: %v", src, err)
        }
        dest := filepath.Join(dir, filepath.Base(src))
        if err := copyFile(src, dest); err != nil {
            t.Fatalf("copy %s: %v", src, err)
        }
        paths = append(paths, dest)
    }

    scanned, err := parser.Scan(dir)
    if err != nil {
        t.Fatalf("scan: %v", err)
    }
    if len(scanned) != len(paths) {
        t.Fatalf("scanned %d, want %d", len(scanned), len(paths))
    }

    st, err := store.Open(filepath.Join(t.TempDir(), "e2e.db"))
    if err != nil {
        t.Fatalf("open: %v", err)
    }
    defer st.Close()

    for _, p := range paths {
        book, err := parser.ParseByExtension(p)
        if err != nil {
            t.Fatalf("parse %s: %v", p, err)
        }
        book.FilePath = p
        if _, err := st.UpsertBook(*book); err != nil {
            t.Fatalf("upsert: %v", err)
        }
    }
    books, _ := st.ListBooks()
    if len(books) != len(paths) {
        t.Fatalf("books: %d, want %d", len(books), len(paths))
    }

    bid := books[0].ID
    if err := st.SaveProgress(models.ReadingProgress{BookID: bid, Chapter: 1, Page: 2}); err != nil {
        t.Fatalf("save progress: %v", err)
    }
    p, _ := st.GetProgress(bid)
    if p.Chapter != 1 || p.Page != 2 {
        t.Fatalf("progress: %+v", p)
    }

    bmID, err := st.AddBookmark(models.Bookmark{BookID: bid, Chapter: 0, Page: 1, Label: "test"})
    if err != nil {
        t.Fatalf("add bookmark: %v", err)
    }
    bms, _ := st.ListBookmarks(bid)
    if len(bms) != 1 || bms[0].ID != bmID {
        t.Fatalf("bookmarks: %+v", bms)
    }

    if err := st.DeleteBook(bid); err != nil {
        t.Fatalf("delete: %v", err)
    }
    bms, _ = st.ListBookmarks(bid)
    if len(bms) != 0 {
        t.Fatalf("expected cascade, got %d", len(bms))
    }
}

func copyFile(src, dst string) error {
    data, err := readFile(src)
    if err != nil {
        return err
    }
    return writeFile(dst, data)
}
```

- [ ] **Step 3: 运行测试**

```bash
go test ./e2e/... -v
```

Expected: PASS

- [ ] **Step 4: 提交**

```bash
git add e2e/
git commit -m "test: 端到端测试（扫描→解析→存储→阅读→书签）"
```

---

## Task 15: 跨平台构建脚本和 README

**Files:**
- Create: `build.sh`
- Create: `README.md`

- [ ] **Step 1: 编写 `build.sh`**

```bash
#!/bin/bash
set -e

APP_NAME="novel-reader"
LDFLAGS="-s -w"

echo "=== 当前平台 ==="
go build -ldflags "$LDFLAGS" -o "${APP_NAME}" .

echo "=== Windows (amd64) ==="
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "${APP_NAME}-windows-amd64.exe" .

echo "=== macOS (amd64) ==="
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "${APP_NAME}-darwin-amd64" .

echo "=== macOS (arm64) ==="
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "${APP_NAME}-darwin-arm64" .

echo "=== 完成 ==="
ls -la ${APP_NAME}* 2>/dev/null || ls -la
```

```bash
chmod +x build.sh
```

- [ ] **Step 2: 编写 `README.md`**

```markdown
# Novel Reader — 终端小说阅读器

一个跨平台（Windows / macOS）的终端小说阅读器，支持 EPUB、TXT、Markdown 格式。

## 功能

- 📚 书架：扫描指定目录下的所有书籍
- 📖 阅读：按终端高度分页，支持中文
- 📑 目录：按章跳转
- 🔖 书签：随时添加、查看、删除
- 💾 进度：自动保存，下次打开继续
- ⌨️ Vim 风格快捷键

## 安装

从源码编译：

\`\`\`bash
git clone <repo>
cd cli-read
go build -o reader .
\`\`\`

或使用构建脚本生成多平台二进制：

\`\`\`bash
./build.sh
\`\`\`

## 使用

\`\`\`bash
./reader
./reader --dir /path/to/books
./reader --db /path/to/db.sqlite
\`\`\`

## 键盘快捷键

### 书架

| 键 | 功能 |
|----|------|
| `j` / `↓` | 向下移动 |
| `k` / `↑` | 向上移动 |
| `Enter` | 打开选中书籍 |
| `/` | 搜索书名 |
| `q` | 退出 |

### 阅读

| 键 | 功能 |
|----|------|
| `j` / `Space` | 下一页 |
| `k` | 上一页 |
| `gg` | 跳到章节开头 |
| `G` | 跳到章节末尾 |
| `o` | 打开章节目录 |
| `n` | 下一章 |
| `p` | 上一章 |
| `m` | 添加书签 |
| `b` | 查看书签 |
| `q` / `Esc` | 返回书架 |

## 支持的格式

- **TXT**: 自动检测 UTF-8 / GBK / GB18030 编码
- **Markdown**: 按 `#` / `##` 标题切分章节
- **EPUB**: 解析 OPF 和 HTML 章节

## 数据存储

所有数据保存在一个 SQLite 文件中（默认 `./novel-reader.db`）：

- `books` — 书架
- `reading_progress` — 每本书的阅读进度
- `bookmarks` — 书签

## 限制

- 不做实时文件监控，新增/删除书籍需重启程序
- 不做云同步
- 不做复杂 Markdown 渲染（表格、代码块）

## 许可

MIT
```

> 注意：上面 README 里的 `\`\`\`` 实际写文件时用普通三反引号即可

- [ ] **Step 3: 验证构建脚本**

```bash
bash build.sh
```

Expected: 当前平台、Windows、macOS（Intel 和 Apple Silicon）二进制文件均生成

- [ ] **Step 4: 验证所有测试**

```bash
go test ./... -v
```

Expected: 所有包测试 PASS

- [ ] **Step 5: 提交**

```bash
git add build.sh README.md
git commit -m "docs: README 和跨平台构建脚本"
```

---

## Self-Review

**1. Spec coverage:**

| Spec 要求 | 覆盖任务 |
|----------|---------|
| Go 语言、跨平台 | Task 1、Task 15 |
| EPUB/TXT/MD 解析 | Task 5、6、7、8 |
| 监控目录扫描 | Task 9、Task 13 |
| SQLite 存储 | Task 2 |
| Vim 风格快捷键 | Task 10、11、12 |
| 分页阅读 | Task 3、Task 12 |
| 跟随终端主题 | Task 10 |
| 章节目录 | Task 12（viewTOC）|
| 阅读进度记录 | Task 2、Task 12（saveProgress）|
| 书架功能 | Task 11、Task 13 |
| 书签功能 | Task 12（viewBookmarks、addBookmark）|
| 编码自动检测 | Task 6 |
| TDD | 每个 Task 都有测试先行的步骤 |

**2. Placeholder scan:** Task 4 的占位文件（epub.go/txt.go/markdown.go 临时返回空 Book）已被后续 Task 5-8 覆盖；Task 1 Step 6 的临时 main.go 骨架被 Task 13 覆盖。无残留 TBD/TODO。

**3. Type consistency:**
- `models.Book/Chapter/ReadingProgress/Bookmark` 字段名一致
- `pager.New/Page/PageCount/LineWidth` 签名一致
- `store.*` 函数签名一致
- `app.NewBookshelfModel/Update/View/Selected/SetStatus` 一致
- `app.NewReaderModel/Update/View` 一致

**4. Import 注意事项：**
- Task 11、12 使用 `key.Matches`，需要 `import "github.com/charmbracelet/bubbles/key"`，代码中已包含
- Task 7、8、6 的测试使用 `strings.Contains`，已显式 import

**5. 边界处理：**
- Task 12 中 `nextPage` 翻到最后一页自动跨章节
- Task 12 中 `resize` 重新计算 pager 边界
- Task 6 中编码检测失败返回错误
- Task 13 中删除数据库中不存在的书籍记录不报错
- Task 8 中 EPUB 章节无 HTML 文件时跳过
