package pager

import (
	"fmt"
	"strings"
)

// Pager splits text into pages by width and height.
// width = terminal columns, height = body lines (excluding status bar).
// CJK characters count as 2 columns.
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

// wrap breaks a line at the given width.
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

// New creates a Pager.
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

// PageCount returns total pages (at least 1).
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

// Page returns the content for page idx (0-based).
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

// LineWidth returns the configured line width.
func (p *Pager) LineWidth() int {
	return p.width
}
