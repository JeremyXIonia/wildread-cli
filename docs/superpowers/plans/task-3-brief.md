# Task 3: Pager 层（文本分页）

**Goal:** Create a Pager that splits text into pages by terminal width/height, with CJK support (Chinese chars = 2 columns).

**Pre-requisites:** None (standalone module)

**Output files:**
- `pager/pager.go`
- `pager/pager_test.go`

## Steps

### Step 1: Create `pager/pager_test.go`
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

### Step 2: Run tests (should FAIL)
`go test ./pager/...`

### Step 3: Create `pager/pager.go`
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
    if r < 0x80 { return 1 }
    return 2
}

func displayWidth(s string) int {
    w := 0
    for _, r := range s { w += runeWidth(r) }
    return w
}

func wrap(line string, width int) []string {
    if displayWidth(line) <= width { return []string{line} }
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
    if len(cur) > 0 { out = append(out, string(cur)) }
    return out
}

func New(text string, width, height int) *Pager {
    if width <= 0 { width = 80 }
    if height <= 0 { height = 24 }
    raw := strings.Split(text, "\n")
    var lines []string
    for _, l := range raw {
        l = strings.TrimRight(l, "\r")
        if l == "" { lines = append(lines, ""); continue }
        lines = append(lines, wrap(l, width)...)
    }
    return &Pager{lines: lines, width: width, pageSize: height}
}

func (p *Pager) PageCount() int {
    if len(p.lines) == 0 { return 1 }
    n := len(p.lines) / p.pageSize
    if len(p.lines)%p.pageSize != 0 { n++ }
    if n < 1 { n = 1 }
    return n
}

func (p *Pager) Page(idx int) (string, error) {
    n := p.PageCount()
    if idx < 0 || idx >= n {
        return "", fmt.Errorf("page %d out of range [0, %d)", idx, n)
    }
    start := idx * p.pageSize
    end := start + p.pageSize
    if end > len(p.lines) { end = len(p.lines) }
    return strings.Join(p.lines[start:end], "\n"), nil
}

func (p *Pager) LineWidth() int { return p.width }
```

### Step 4: Run tests (should PASS)
`go test ./pager/... -v`

### Step 5: Commit
`git add pager/ && git commit -m "feat: pager 层（文本分页，支持中文）"`

## Report
Write report to `docs/superpowers/plans/task-3-report.md`.
