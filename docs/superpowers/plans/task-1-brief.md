# Task 1: 项目骨架与数据模型

**Goal:** Initialize the Go module, add dependencies, define data models, create a skeleton main.go, and initialize git.

**Output files:**
- `go.mod`
- `main.go`
- `models/book.go`
- `models/book_test.go`

## Global Constraints (relevant)
- Go version ≥ 1.22
- Module path: `github.com/xuanchong/cli-read`
- Dependencies: bubbletea, bubbles, lipgloss, go-sqlite3, golang.org/x/text, golang.org/x/net/html, goldmark

## Steps

1. `cd D:\workspace-latest\cli-read && go mod init github.com/xuanchong/cli-read`

2. Add dependencies:
```
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/mattn/go-sqlite3@latest
go get golang.org/x/text@latest
go get golang.org/x/net/html@latest
go get github.com/yuin/goldmark@latest
```

3. Create `models/book.go`:
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

4. Create `models/book_test.go`:
```go
package models

import "testing"

func TestBookFields(t *testing.T) {
    b := Book{ID: 1, FilePath: "/x.epub", Title: "X", Format: "epub"}
    if b.ID != 1 || b.Title != "X" || b.Format != "epub" {
        t.Fatalf("unexpected book: %+v", b)
    }
}

func TestChapterContent(t *testing.T) {
    c := Chapter{Title: "ch1", Content: "p1\n\np2"}
    if c.Title != "ch1" || c.Content != "p1\n\np2" {
        t.Fatalf("unexpected chapter: %+v", c)
    }
}
```

5. `go test ./models/...` — should PASS

6. Create `main.go`:
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

7. `go build -o reader.exe .` — should succeed

8. `mkdir -p ./books && ./reader.exe` — should print paths and exit cleanly

9. `git init`

10. `git add go.mod go.sum main.go models/ && git commit -m "feat: 项目骨架与数据模型"`

## Report
Write report to `docs/superpowers/plans/task-1-report.md` with:
- Status (DONE/NEEDS_CONTEXT/BLOCKED)
- Commits (commit hash)
- Test results
- Any concerns
