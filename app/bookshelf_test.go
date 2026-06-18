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
