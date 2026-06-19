package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeremyXIonia/wildread-cli/models"
	"github.com/JeremyXIonia/wildread-cli/parser"
	"github.com/JeremyXIonia/wildread-cli/store"
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
}

func TestReaderPreservesProgressPageUntilResize(t *testing.T) {
	var lines []string
	for i := 0; i < 120; i++ {
		lines = append(lines, fmt.Sprintf("%03d %s", i, strings.Repeat("字", 20)))
	}
	book := &models.Book{
		ID:    1,
		Title: "long",
		Chapters: []models.Chapter{{
			Title:   "chapter",
			Content: strings.Join(lines, "\n"),
		}},
	}

	m := NewReaderModel(book, models.ReadingProgress{BookID: 1, Chapter: 0, Page: 20}, nil)
	if m.page != 20 {
		t.Fatalf("page was reset before resize: got %d, want 20", m.page)
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 20, Height: 8})
	rm := updated.(ReaderModel)
	if rm.page != 20 {
		t.Fatalf("page after small resize: got %d, want 20", rm.page)
	}
}

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
