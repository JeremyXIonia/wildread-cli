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
