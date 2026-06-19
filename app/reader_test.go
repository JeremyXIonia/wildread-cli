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

func TestReaderTOCSelectionJumpsToSelectedChapter(t *testing.T) {
	book := &models.Book{
		ID:    1,
		Title: "chapters",
		Chapters: []models.Chapter{
			{Title: "第一章", Content: "one"},
			{Title: "第二章", Content: "two"},
		},
	}
	m := NewReaderModel(book, models.ReadingProgress{BookID: 1}, nil)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	rm := updated.(ReaderModel)

	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm = updated.(ReaderModel)
	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	rm = updated.(ReaderModel)

	if rm.mode != ModeReading {
		t.Fatalf("mode = %d, want ModeReading", rm.mode)
	}
	if rm.chapter != 1 || rm.page != 0 {
		t.Fatalf("location = chapter %d page %d, want chapter 1 page 0", rm.chapter, rm.page)
	}
	if !strings.Contains(rm.View(), "第二章") {
		t.Fatalf("reader did not jump to selected chapter: %s", rm.View())
	}
}

func TestReaderBookmarksSelectionJumpsToSelectedBookmark(t *testing.T) {
	book := &models.Book{
		ID:    1,
		Title: "bookmarks",
		Chapters: []models.Chapter{
			{Title: "第一章", Content: strings.Repeat("one\n", 80)},
			{Title: "第二章", Content: strings.Repeat("two\n", 80)},
		},
	}
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	bookID, err := st.UpsertBook(models.Book{FilePath: "/bookmarks.md", Title: book.Title, Format: "md"})
	if err != nil {
		t.Fatalf("upsert book: %v", err)
	}
	book.ID = bookID
	if _, err := st.AddBookmark(models.Bookmark{BookID: bookID, Chapter: 0, Page: 0, Label: "first"}); err != nil {
		t.Fatalf("add first bookmark: %v", err)
	}
	if _, err := st.AddBookmark(models.Bookmark{BookID: bookID, Chapter: 1, Page: 2, Label: "second"}); err != nil {
		t.Fatalf("add second bookmark: %v", err)
	}
	m := NewReaderModel(book, models.ReadingProgress{BookID: bookID}, st)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	rm := updated.(ReaderModel)
	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	rm = updated.(ReaderModel)

	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	rm = updated.(ReaderModel)
	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	rm = updated.(ReaderModel)

	if rm.mode != ModeReading {
		t.Fatalf("mode = %d, want ModeReading", rm.mode)
	}
	if rm.chapter != 1 || rm.page != 2 {
		t.Fatalf("location = chapter %d page %d, want chapter 1 page 2", rm.chapter, rm.page)
	}
}

func TestReaderTOCViewKeepsSelectedChapterVisibleInLongList(t *testing.T) {
	var chapters []models.Chapter
	for i := 0; i < 30; i++ {
		chapters = append(chapters, models.Chapter{Title: fmt.Sprintf("第 %02d 章", i+1), Content: "content"})
	}
	book := &models.Book{ID: 1, Title: "long toc", Chapters: chapters}
	m := NewReaderModel(book, models.ReadingProgress{BookID: 1, Chapter: 20}, nil)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	rm := updated.(ReaderModel)
	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	rm = updated.(ReaderModel)

	view := rm.View()
	if strings.Count(view, "\n")+1 > 10 {
		t.Fatalf("toc view exceeds terminal height: lines=%d view=%q", strings.Count(view, "\n")+1, view)
	}
	if !strings.Contains(view, "> 第 21 章") {
		t.Fatalf("toc view does not keep selected chapter visible: %q", view)
	}
}

func TestReaderBookmarksViewKeepsSelectedBookmarkVisibleInLongList(t *testing.T) {
	book := &models.Book{ID: 1, Title: "long bookmarks", Chapters: []models.Chapter{{Title: "章", Content: strings.Repeat("content\n", 80)}}}
	st, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	bookID, err := st.UpsertBook(models.Book{FilePath: "/long-bookmarks.md", Title: book.Title, Format: "md"})
	if err != nil {
		t.Fatalf("upsert book: %v", err)
	}
	book.ID = bookID
	for i := 0; i < 30; i++ {
		if _, err := st.AddBookmark(models.Bookmark{BookID: bookID, Chapter: 0, Page: i, Label: fmt.Sprintf("bookmark-%02d", i+1)}); err != nil {
			t.Fatalf("add bookmark %d: %v", i, err)
		}
	}
	m := NewReaderModel(book, models.ReadingProgress{BookID: bookID}, st)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	rm := updated.(ReaderModel)
	updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	rm = updated.(ReaderModel)
	for i := 0; i < 20; i++ {
		updated, _ = rm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		rm = updated.(ReaderModel)
	}

	view := rm.View()
	if strings.Count(view, "\n")+1 > 10 {
		t.Fatalf("bookmarks view exceeds terminal height: lines=%d view=%q", strings.Count(view, "\n")+1, view)
	}
	if !strings.Contains(view, "> #") || !strings.Contains(view, "bookmark-21") {
		t.Fatalf("bookmarks view does not keep selected bookmark visible: %q", view)
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
