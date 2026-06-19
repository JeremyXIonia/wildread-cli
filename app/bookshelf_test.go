package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeremyXIonia/wildread-cli/models"
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

func TestBookshelfOpenDirectoryManagerWithLowercaseDSendsMsg(t *testing.T) {
	books := []models.Book{{ID: 1, Title: "三体", Format: "epub"}}
	m := NewBookshelfModel(books)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected command")
	}
	if _, ok := cmd().(OpenDirectoryManagerMsg); !ok {
		t.Fatalf("expected OpenDirectoryManagerMsg")
	}
}

func TestBookshelfViewShowsDirectoryManagerHintInListHelp(t *testing.T) {
	books := []models.Book{{ID: 1, Title: "三体", Format: "epub"}}
	m := NewBookshelfModel(books)
	view := m.View()
	if !strings.Contains(view, "d 目录") {
		t.Fatalf("view missing directory manager hint in list help: %q", view)
	}
	if strings.Contains(view, "d 管理目录") {
		t.Fatalf("view still renders separate directory manager hint: %q", view)
	}
}
