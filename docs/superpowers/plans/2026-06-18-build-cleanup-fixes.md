# Build Cleanup Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the Windows-specific one-click build script and fix the related build, binary, reader progress, bookmark return, and initialization issues found in review.

**Architecture:** Keep the project as a standard cross-platform Go application: local development uses `go test ./...` and `go build`, while release-style multi-target builds use the existing POSIX `build.sh`. Reader state fixes stay inside `app.ReaderModel`, while app-level back navigation is delegated to the reader before falling back to the bookshelf.

**Tech Stack:** Go, Bubble Tea, Bubbles viewport, SQLite via `modernc.org/sqlite`, shell build script.

---

## File Structure

- Delete: `build.bat` — remove Windows-specific one-click build script so development is not centered on Windows batch tooling.
- Modify: `README.md` — document native Go commands as the primary development/build path and remove `build.bat` references.
- Modify: `CLAUDE.md` — replace Windows-specific build/run path with cross-platform Go commands.
- Modify: `build.sh` — restore stripped linker flags for release artifacts.
- Modify: `go.mod` — align `go` directive with README's Go 1.22+ requirement.
- Modify: `app/reader.go` — preserve saved page progress until actual terminal size is known, remove duplicate `Init()` page load, expose whether reader is in the main reading mode.
- Modify: `main.go` — let reader submodes handle `q`/`esc` before root returns to bookshelf.
- Modify: `app/reader_test.go` — add regression tests for saved progress preservation and reader submode state.
- Modify git index: stop tracking `reader.exe` while leaving it ignored by `.gitignore`.

---

### Task 1: Remove Windows batch build path from docs and files

**Files:**
- Delete: `build.bat`
- Modify: `README.md:19-59`
- Modify: `CLAUDE.md:3-13`

- [ ] **Step 1: Delete the Windows-only build script**

Run:
```bash
rm build.bat
```

Expected: `git status --short` shows `D build.bat` if it was tracked, or the untracked file disappears if it was untracked.

- [ ] **Step 2: Update README build section**

Replace the README build section with text equivalent to:

```markdown
## 编译

### 日常开发

```bash
# 运行测试
go test ./...

# 编译当前平台
go build .
```

如需指定输出文件名：

```bash
# macOS / Linux
go build -o reader .

# Windows
go build -o reader.exe .
```

### 交叉编译示例

```bash
# 在 macOS / Linux 上编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -o reader.exe .

# 在 Windows CMD 中编译 macOS Apple Silicon 版本
set GOOS=darwin
set GOARCH=arm64
go build -o reader-mac-arm64 .
```

### 一键构建发布产物

```bash
./build.sh        # macOS / Linux / Git Bash / WSL
```

生成当前平台、Windows (amd64)、macOS (Intel + Apple Silicon) 的二进制文件。
```

- [ ] **Step 3: Update CLAUDE.md build instructions**

Change the build block to:

```markdown
## Build & Run
```bash
go build -o reader .
./reader --dir ./books
```

Windows:
```powershell
go build -o reader.exe .
.\reader.exe --dir .\books
```
```

Keep the existing test command:

```bash
go test ./... -v
```

- [ ] **Step 4: Verify no docs mention build.bat**

Run:
```bash
grep -R "build\.bat" -n README.md CLAUDE.md docs || true
```

Expected: no output.

---

### Task 2: Align Go version and release build size

**Files:**
- Modify: `go.mod:3`
- Modify: `build.sh:1-19`

- [ ] **Step 1: Change go.mod version**

Change:
```go
go 1.26.2
```

to:
```go
go 1.22
```

- [ ] **Step 2: Restore stripped release flags in build.sh**

Change `build.sh` to:

```bash
#!/bin/bash
set -e

LDFLAGS="-s -w"

mkdir -p bin

echo "=== 当前平台 ==="
go build -ldflags "$LDFLAGS" -o reader .

echo "=== Windows (amd64) ==="
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/reader-windows-amd64.exe .

echo "=== macOS (amd64 Intel) ==="
GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/reader-darwin-amd64 .

echo "=== macOS (arm64 Apple Silicon) ==="
GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o bin/reader-darwin-arm64 .

echo "=== 完成 ==="
ls -la bin/reader-*
```

- [ ] **Step 3: Verify module and shell script syntax**

Run:
```bash
go mod tidy
bash -n build.sh
```

Expected: both commands succeed. `go.mod` remains at `go 1.22`.

---

### Task 3: Stop tracking generated binary

**Files:**
- Git index only: `reader.exe`
- Existing ignore rules: `.gitignore:1-5`

- [ ] **Step 1: Remove reader.exe from git tracking without deleting local file**

Run:
```bash
git rm --cached reader.exe
```

Expected: `git status --short` shows `D  reader.exe`, and local `reader.exe` still exists on disk.

- [ ] **Step 2: Verify ignore rules already cover future binaries**

Run:
```bash
git check-ignore -v reader.exe reader bin/reader-windows-amd64.exe
```

Expected: output points to `.gitignore` rules for `*.exe` or `reader*`.

---

### Task 4: Preserve saved page progress until real terminal sizing

**Files:**
- Modify: `app/reader.go:47-90`, `app/reader.go:238-257`
- Modify: `app/reader_test.go`

- [ ] **Step 1: Add failing test for saved page preservation**

Append to `app/reader_test.go`:

```go
func TestReaderPreservesProgressPageUntilResize(t *testing.T) {
	book := &models.Book{
		ID:    1,
		Title: "long",
		Chapters: []models.Chapter{{
			Title:   "chapter",
			Content: strings.Repeat("line\n", 120),
		}},
	}

	m := NewReaderModel(book, models.ReadingProgress{BookID: 1, Chapter: 0, Page: 30}, nil)
	if m.page != 30 {
		t.Fatalf("page was reset before resize: got %d, want 30", m.page)
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 20, Height: 8})
	rm := updated.(ReaderModel)
	if rm.page != 30 {
		t.Fatalf("page after small resize: got %d, want 30", rm.page)
	}
}
```

Add `strings` to the imports in `app/reader_test.go`:

```go
import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	...
)
```

- [ ] **Step 2: Run the failing test**

Run:
```bash
go test ./app -run TestReaderPreservesProgressPageUntilResize -v
```

Expected before implementation: FAIL with `page was reset before resize`.

- [ ] **Step 3: Update NewReaderModel clamping**

In `app/reader.go`, change the progress normalization to clamp negative values but do not reset high page values before real sizing.

Use this logic near the top of `NewReaderModel`:

```go
chapter := progress.Chapter
page := progress.Page
if len(book.Chapters) == 0 {
	book.Chapters = []models.Chapter{{Title: "空", Content: "（无内容）"}}
}
if chapter < 0 || chapter >= len(book.Chapters) {
	chapter = 0
}
if page < 0 {
	page = 0
}
```

Remove this pre-resize reset:

```go
if page >= p.PageCount() {
	page = 0
}
```

- [ ] **Step 4: Make initial content loading tolerate an out-of-range page**

Update `loadPageContent` to clamp against the current pager only when content is actually loaded:

```go
func (m *ReaderModel) loadPageContent() {
	if m.page >= m.pager.PageCount() {
		m.page = m.pager.PageCount() - 1
	}
	if m.page < 0 {
		m.page = 0
	}
	content, err := m.pager.Page(m.page)
	if err != nil {
		m.status = "页码错误"
		return
	}
	m.viewport.SetContent(content)
	m.status = ""
}
```

- [ ] **Step 5: Remove eager content load from constructor**

Delete this line from `NewReaderModel`:

```go
m.loadPageContent()
```

Keep `viewport` empty until either `Init` or `WindowSizeMsg` loads content.

- [ ] **Step 6: Remove duplicate discarded load from Init**

Change:

```go
func (m ReaderModel) Init() tea.Cmd {
	m.loadPageContent()
	return m.saveProgress()
}
```

to:

```go
func (m ReaderModel) Init() tea.Cmd {
	return nil
}
```

This avoids saving a clamped page before the real terminal size is known.

- [ ] **Step 7: Ensure resize clamps using real size**

Keep `resize()` rebuilding the pager from `bodyWidth` and `bodyHeight`, then calling `m.loadPageContent()`:

```go
m.pager = pager.New(m.book.Chapters[m.chapter].Content, bodyWidth, bodyHeight)
m.loadPageContent()
```

Remove duplicate clamp code in `resize()` if `loadPageContent()` now owns clamping.

- [ ] **Step 8: Run reader tests**

Run:
```bash
go test ./app -v
```

Expected: PASS.

---

### Task 5: Let reader submodes handle q/esc before leaving reader

**Files:**
- Modify: `app/reader.go`
- Modify: `main.go:148-165`
- Modify: `app/reader_test.go`

- [ ] **Step 1: Add reader mode helper**

Add to `app/reader.go` after `Init()`:

```go
// IsReading reports whether the reader is showing the main reading view.
func (m ReaderModel) IsReading() bool {
	return m.mode == ModeReading
}
```

- [ ] **Step 2: Add reader submode back test**

Append to `app/reader_test.go`:

```go
func TestReaderBookmarksBackReturnsToReading(t *testing.T) {
	m, _ := newTestReader(t)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	rm := updated.(ReaderModel)
	if rm.mode != ModeBookmarks {
		t.Fatalf("expected bookmarks mode, got %d", rm.mode)
	}

	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	rm = updated.(ReaderModel)
	if rm.mode != ModeReading {
		t.Fatalf("expected reading mode after esc, got %d", rm.mode)
	}
}
```

- [ ] **Step 3: Add root-level delegation test if practical**

If root tests are introduced, create `main_test.go` with a test that sets `rootModel{mode: modeReader, reader: &readerModelInBookmarkMode}` and verifies `esc` stays in `modeReader`. If direct construction is awkward because `ReaderModel.mode` is unexported, skip root-specific construction and rely on the `IsReading` behavior in Step 4.

- [ ] **Step 4: Update root key interception**

In `main.go`, change:

```go
if m.mode == modeReader && (msg.String() == "esc" || msg.String() == "q") {
	m.mode = modeBookshelf
	return m, nil
}
```

to:

```go
if m.mode == modeReader && m.reader != nil && m.reader.IsReading() && (msg.String() == "esc" || msg.String() == "q") {
	m.mode = modeBookshelf
	return m, nil
}
```

This allows TOC/bookmark submodes to receive `q`/`esc` and return to reading.

- [ ] **Step 5: Run app tests**

Run:
```bash
go test ./app -v
```

Expected: PASS.

---

### Task 6: Final verification

**Files:**
- All modified files

- [ ] **Step 1: Format Go files**

Run:
```bash
gofmt -w main.go app/reader.go app/reader_test.go
```

Expected: no output.

- [ ] **Step 2: Run full tests**

Run:
```bash
go test ./... -v
```

Expected: PASS.

- [ ] **Step 3: Check final diff**

Run:
```bash
git diff --stat
git status --short
```

Expected:
- `build.bat` removed from untracked files.
- `reader.exe` removed from tracking.
- Source/docs changes are reviewable text diffs.
- No unexpected generated files are staged or modified.

---

## Self-Review

- Spec coverage:方案 B deletion of `build.bat`, non-Windows-centric docs, Go version mismatch, build artifact size, tracked binary, reader progress reset, bookmark back handling, and duplicate initialization are each covered by a task.
- Placeholder scan: no TBD/TODO placeholders remain; every task has concrete commands or code.
- Type consistency: `ReaderModel.IsReading()` is defined before `main.go` uses it; tests use existing `models.Book`, `models.Chapter`, `tea.KeyMsg`, and `ReaderModel` internals from package `app` tests.
