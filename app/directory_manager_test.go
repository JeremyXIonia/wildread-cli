package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeremyXIonia/wildread-cli/models"
)

func testDirs() []models.LibraryDir {
	return []models.LibraryDir{
		{ID: 1, Path: "/tmp/.cli-read/.book", IsDefault: true},
		{ID: 2, Path: "/tmp/books", IsDefault: false},
	}
}

func TestDirectoryManagerViewListsDirs(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	view := m.View()
	if !strings.Contains(view, "书籍目录") || !strings.Contains(view, "/tmp/books") || !strings.Contains(view, "默认") {
		t.Fatalf("view missing content: %s", view)
	}
}

func TestDirectoryManagerAddEmitsAddMessage(t *testing.T) {
	path := t.TempDir()
	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(DirectoryManagerModel)
	if m.mode != dirModeAdd {
		t.Fatalf("mode: %d", m.mode)
	}
	for _, r := range path {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(DirectoryManagerModel)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected add command")
	}
	msg := cmd()
	add, ok := msg.(AddLibraryDirMsg)
	if !ok {
		t.Fatalf("message: %T", msg)
	}
	if add.Path != path || add.Create {
		t.Fatalf("add msg: %+v", add)
	}
}

func TestDirectoryManagerRejectsExistingRegularFileAsLibraryDir(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "not-a-dir.txt")
	if err := os.WriteFile(filePath, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(DirectoryManagerModel)
	for _, r := range filePath {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(DirectoryManagerModel)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(DirectoryManagerModel)
	if cmd != nil {
		t.Fatalf("expected no add command for regular file")
	}
	if m.mode != dirModeList {
		t.Fatalf("mode: %d", m.mode)
	}
	view := m.View()
	want := "不是目录: " + filePath
	if !strings.Contains(view, want) {
		t.Fatalf("view missing %q: %s", want, view)
	}
}

func TestDirectoryManagerAddEmptyInputStatusIsVisible(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(DirectoryManagerModel)

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(DirectoryManagerModel)
	if cmd != nil {
		t.Fatalf("expected no command for empty input")
	}
	if m.mode != dirModeAdd {
		t.Fatalf("mode: %d", m.mode)
	}
	view := m.View()
	if !strings.Contains(view, "目录不能为空") {
		t.Fatalf("add view missing empty-input status: %s", view)
	}
}

func TestDirectoryManagerAddMissingPathAsksWhetherToCreate(t *testing.T) {
	path := t.TempDir() + "/new-books"
	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(DirectoryManagerModel)
	for _, r := range path {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(DirectoryManagerModel)
	}

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(DirectoryManagerModel)
	if cmd != nil {
		t.Fatalf("expected no add command before create confirmation")
	}
	if m.mode != dirModeConfirmCreate {
		t.Fatalf("mode: %d", m.mode)
	}
	view := m.View()
	if !strings.Contains(view, "目录不存在") || !strings.Contains(view, "创建") || !strings.Contains(view, path) {
		t.Fatalf("confirm create view: %s", view)
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected add create command")
	}
	msg := cmd()
	add, ok := msg.(AddLibraryDirMsg)
	if !ok {
		t.Fatalf("message: %T", msg)
	}
	if add.Path != path || !add.Create {
		t.Fatalf("add create msg: %+v", add)
	}
}

func TestDirectoryManagerDeleteRequiresConfirmation(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = updated.(DirectoryManagerModel)
	if m.mode != dirModeConfirmDelete {
		t.Fatalf("mode: %d", m.mode)
	}
	view := m.View()
	if !strings.Contains(view, "阅读进度和书签") || !strings.Contains(view, "原始文件不会被删除") {
		t.Fatalf("confirm view: %s", view)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatal("expected delete command")
	}
	msg := cmd()
	del, ok := msg.(DeleteLibraryDirMsg)
	if !ok {
		t.Fatalf("message: %T", msg)
	}
	if del.Dir.ID != 1 {
		t.Fatalf("delete msg: %+v", del)
	}
}

func TestDirectoryManagerBackEmitsClose(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected close command")
	}
	if _, ok := cmd().(CloseDirectoryManagerMsg); !ok {
		t.Fatalf("message: %T", cmd())
	}
}

func TestDirectoryManagerRescanEmitsMessage(t *testing.T) {
	m := NewDirectoryManagerModel(testDirs())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected rescan command")
	}
	if _, ok := cmd().(RescanLibraryDirsMsg); !ok {
		t.Fatalf("message: %T", cmd())
	}
}
