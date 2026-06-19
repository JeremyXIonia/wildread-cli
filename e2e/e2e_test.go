package e2e

import (
	"path/filepath"
	"testing"

	"github.com/JeremyXIonia/wildread-cli/models"
	"github.com/JeremyXIonia/wildread-cli/parser"
	"github.com/JeremyXIonia/wildread-cli/store"
)

func TestE2EFlow(t *testing.T) {
	dir := t.TempDir()

	sources := []string{
		filepath.Join("..", "testdata", "sample.txt"),
		filepath.Join("..", "testdata", "sample.md"),
		filepath.Join("..", "testdata", "sample.epub"),
	}

	var paths []string
	for _, src := range sources {
		if _, err := readFile(src); err != nil {
			t.Skipf("missing source %s: %v", src, err)
		}
		dest := filepath.Join(dir, filepath.Base(src))
		if err := copyFile(src, dest); err != nil {
			t.Fatalf("copy %s: %v", src, err)
		}
		paths = append(paths, dest)
	}

	// 1. Scan
	scanned, err := parser.Scan(dir)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(scanned) != len(paths) {
		t.Fatalf("scanned %d, want %d", len(scanned), len(paths))
	}

	// 2. Sync to store
	st, err := store.Open(filepath.Join(t.TempDir(), "e2e.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()

	for _, p := range paths {
		book, err := parser.ParseByExtension(p)
		if err != nil {
			t.Fatalf("parse %s: %v", p, err)
		}
		book.FilePath = p
		if _, err := st.UpsertBook(*book); err != nil {
			t.Fatalf("upsert: %v", err)
		}
	}

	books, _ := st.ListBooks()
	if len(books) != len(paths) {
		t.Fatalf("books: %d, want %d", len(books), len(paths))
	}

	// 3. Reading progress
	bid := books[0].ID
	if err := st.SaveProgress(models.ReadingProgress{BookID: bid, Chapter: 1, Page: 2}); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	p, _ := st.GetProgress(bid)
	if p.Chapter != 1 || p.Page != 2 {
		t.Fatalf("progress: %+v", p)
	}

	// 4. Bookmarks
	bmID, err := st.AddBookmark(models.Bookmark{BookID: bid, Chapter: 0, Page: 1, Label: "test"})
	if err != nil {
		t.Fatalf("add bookmark: %v", err)
	}
	bms, _ := st.ListBookmarks(bid)
	if len(bms) != 1 || bms[0].ID != bmID {
		t.Fatalf("bookmarks: %+v", bms)
	}

	// 5. Cascade delete
	if err := st.DeleteBook(bid); err != nil {
		t.Fatalf("delete: %v", err)
	}
	bms, _ = st.ListBookmarks(bid)
	if len(bms) != 0 {
		t.Fatalf("expected cascade, got %d", len(bms))
	}
}

func TestManagedDirectoryDeleteRemovesBooksAndProgress(t *testing.T) {
	dir := t.TempDir()
	bookPath := filepath.Join(dir, "managed.txt")
	if err := writeFile(bookPath, []byte("Managed\n\n第一章\n内容")); err != nil {
		t.Fatalf("write book: %v", err)
	}

	st, err := store.Open(filepath.Join(t.TempDir(), "e2e.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()

	id, err := st.AddLibraryDir(dir, true)
	if err != nil {
		t.Fatalf("add dir: %v", err)
	}
	book, err := parser.ParseByExtension(bookPath)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	book.FilePath = bookPath
	bookID, err := st.UpsertBook(*book)
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := st.SaveProgress(models.ReadingProgress{BookID: bookID, Chapter: 1, Page: 2}); err != nil {
		t.Fatalf("progress: %v", err)
	}
	if _, err := st.AddBookmark(models.Bookmark{BookID: bookID, Chapter: 0, Page: 0, Label: "mark"}); err != nil {
		t.Fatalf("bookmark: %v", err)
	}
	if err := st.DeleteLibraryDir(id); err != nil {
		t.Fatalf("delete dir: %v", err)
	}
	books, err := st.ListBooks()
	if err != nil {
		t.Fatalf("books: %v", err)
	}
	if len(books) != 0 {
		t.Fatalf("books remain: %+v", books)
	}
	progress, err := st.GetProgress(bookID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if progress.Chapter != 0 || progress.Page != 0 {
		t.Fatalf("progress not cascaded: %+v", progress)
	}
	bookmarks, err := st.ListBookmarks(bookID)
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}
	if len(bookmarks) != 0 {
		t.Fatalf("bookmarks not cascaded: %+v", bookmarks)
	}
	if _, err := readFile(bookPath); err != nil {
		t.Fatalf("original file should remain: %v", err)
	}
}

func copyFile(src, dst string) error {
	data, err := readFile(src)
	if err != nil {
		return err
	}
	return writeFile(dst, data)
}
