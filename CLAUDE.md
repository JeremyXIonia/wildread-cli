# CLAUDE.md — 终端小说阅读器

## Build & Run
```bash
cd D:\workspace-latest\cli-read
go build -o reader.exe .
./reader.exe --dir ./books
```

## All Tests
```bash
go test ./... -v
```

## Project Structure
- `main.go` — Entry point, CLI args, TUI main loop
- `app/` — Bubble Tea TUI components (bookshelf, reader)
- `models/` — Data types (Book, Chapter, ReadingProgress, Bookmark)
- `store/` — SQLite persistence layer
- `parser/` — Document parsers (TXT, Markdown, EPUB)
- `pager/` — Text pagination with CJK support
- `ui/` — Key bindings and minimal styles
- `e2e/` — End-to-end tests
- `testdata/` — Sample files for testing

## Key Dependencies
- github.com/charmbracelet/bubbletea — TUI framework
- github.com/charmbracelet/bubbles — UI components (list, viewport, textinput)
- modernc.org/sqlite — Pure Go SQLite (NO CGO!)
- github.com/yuin/goldmark — Markdown parser
- golang.org/x/net/html — HTML parsing (EPUB)
- golang.org/x/text — Encoding detection (GBK/GB18030)

## Key Design Decisions
- **Pure Go SQLite** — `modernc.org/sqlite` replaces `go-sqlite3`, no MSVC/CGO needed
- Vim-style keybindings (j/k/gg/G/o/n/p/m/b/q)
- Auto-detect encoding for TXT files (UTF-8, GBK, GB18030)
- EPUB parsing: self-implemented using archive/zip + XML + HTML parsing
- Bookshelf: scan directory on startup only (no real-time monitoring)
- Data: SQLite single file (books, reading_progress, bookmarks)
- Theme: follows terminal colors (no custom color scheme)
- Foreign keys: enabled via `PRAGMA foreign_keys = ON` (modernc.org/sqlite doesn't support `?_fk=1`)
