package store

import (
	"github.com/xuanchong/cli-read/models"
	"os"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { s.Close(); os.Remove(path) })
	return s
}

func TestOpenClose(t *testing.T) {
	s := newTestStore(t)
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestOpenConstrainsPoolToOneConnectionForConnectionLocalPragmas(t *testing.T) {
	s := newTestStore(t)
	stats := s.db.Stats()
	if stats.MaxOpenConnections != 1 {
		t.Fatalf("MaxOpenConnections = %d, want 1 so every operation uses the foreign_keys-enabled connection", stats.MaxOpenConnections)
	}
}

func TestBooksCRUD(t *testing.T) {
	s := newTestStore(t)
	b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
	id, err := s.UpsertBook(b)
	if err != nil || id <= 0 {
		t.Fatalf("upsert: %v, %d", err, id)
	}
	got, _ := s.GetBook(id)
	if got.Title != "A" || got.Format != "epub" {
		t.Fatalf("bad book: %+v", got)
	}
	list, _ := s.ListBooks()
	if len(list) != 1 {
		t.Fatalf("list: %d", len(list))
	}
	s.DeleteBook(id)
	list, _ = s.ListBooks()
	if len(list) != 0 {
		t.Fatalf("not empty")
	}
}

func TestUpsertBookUpdates(t *testing.T) {
	s := newTestStore(t)
	b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
	id, _ := s.UpsertBook(b)
	b.ID = id
	b.Title = "A-new"
	id2, _ := s.UpsertBook(b)
	if id != id2 {
		t.Fatalf("id mismatch")
	}
	got, _ := s.GetBook(id)
	if got.Title != "A-new" {
		t.Fatalf("not updated")
	}
}

func TestProgressCRUD(t *testing.T) {
	s := newTestStore(t)
	id, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
	p, _ := s.GetProgress(id)
	if p.Chapter != 0 || p.Page != 0 {
		t.Fatalf("default: %+v", p)
	}
	s.SaveProgress(models.ReadingProgress{BookID: id, Chapter: 2, Page: 5})
	p, _ = s.GetProgress(id)
	if p.Chapter != 2 || p.Page != 5 {
		t.Fatalf("saved: %+v", p)
	}
}

func TestBookmarksCRUD(t *testing.T) {
	s := newTestStore(t)
	bid, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
	id1, _ := s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 1, Page: 2, Label: "m1"})
	s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 3, Page: 4, Label: "m2"})
	list, _ := s.ListBookmarks(bid)
	if len(list) != 2 || list[0].Label != "m1" {
		t.Fatalf("list: %+v", list)
	}
	s.DeleteBookmark(id1)
	list, _ = s.ListBookmarks(bid)
	if len(list) != 1 {
		t.Fatalf("after delete: %d", len(list))
	}
}

func TestLibraryDirsCRUD(t *testing.T) {
	st := newTestStore(t)

	id, err := st.AddLibraryDir("/tmp/books", true)
	if err != nil {
		t.Fatalf("add: %v", err)
	}

	dirs, err := st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("dirs len: %d", len(dirs))
	}
	if dirs[0].ID != id || dirs[0].Path != "/tmp/books" || !dirs[0].IsDefault {
		t.Fatalf("dir: %+v", dirs[0])
	}

	if _, err := st.AddLibraryDir("/tmp/books", false); err == nil {
		t.Fatal("expected duplicate error")
	}

	if err := st.DeleteLibraryDir(id); err != nil {
		t.Fatalf("delete: %v", err)
	}
	dirs, err = st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(dirs) != 0 {
		t.Fatalf("dirs after delete: %+v", dirs)
	}
}

func TestDeleteLibraryDirDeletesBooksUnderDirAndCascadesDependents(t *testing.T) {
	st := newTestStore(t)

	dirID, err := st.AddLibraryDir("/library/books_%", false)
	if err != nil {
		t.Fatalf("add library dir: %v", err)
	}
	deleteID, err := st.UpsertBook(models.Book{FilePath: "/library/books_%/sub/a.txt", Title: "delete", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert delete: %v", err)
	}
	rootFileID, err := st.UpsertBook(models.Book{FilePath: "/library/books_%", Title: "delete root", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert root file: %v", err)
	}
	siblingID, err := st.UpsertBook(models.Book{FilePath: "/library/books_%sibling/b.txt", Title: "keep sibling", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert sibling: %v", err)
	}
	outsideID, err := st.UpsertBook(models.Book{FilePath: "/outside/books_%/c.txt", Title: "keep outside", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert outside: %v", err)
	}
	if err := st.SaveProgress(models.ReadingProgress{BookID: deleteID, Chapter: 1, Page: 2}); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if _, err := st.AddBookmark(models.Bookmark{BookID: deleteID, Chapter: 0, Page: 0, Label: "mark"}); err != nil {
		t.Fatalf("bookmark: %v", err)
	}

	if err := st.DeleteLibraryDir(dirID); err != nil {
		t.Fatalf("delete library dir: %v", err)
	}

	dirs, err := st.ListLibraryDirs()
	if err != nil {
		t.Fatalf("list dirs: %v", err)
	}
	if len(dirs) != 0 {
		t.Fatalf("library dir still exists: %+v", dirs)
	}
	if _, err := st.GetBook(deleteID); err == nil {
		t.Fatal("book under deleted library dir still exists")
	}
	if _, err := st.GetBook(rootFileID); err == nil {
		t.Fatal("book matching deleted library dir path still exists")
	}
	if got, err := st.GetBook(siblingID); err != nil || got.ID != siblingID {
		t.Fatalf("sibling book deleted: book=%+v err=%v", got, err)
	}
	if got, err := st.GetBook(outsideID); err != nil || got.ID != outsideID {
		t.Fatalf("outside book deleted: book=%+v err=%v", got, err)
	}
	progress, err := st.GetProgress(deleteID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if progress.Chapter != 0 || progress.Page != 0 {
		t.Fatalf("progress not cascaded: %+v", progress)
	}
	marks, err := st.ListBookmarks(deleteID)
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}
	if len(marks) != 0 {
		t.Fatalf("bookmarks not cascaded: %+v", marks)
	}
}

func TestDeleteBooksUnderDirCascadesProgressAndBookmarks(t *testing.T) {
	st := newTestStore(t)

	keepID, err := st.UpsertBook(models.Book{FilePath: "/books-keep/a.txt", Title: "keep", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert keep: %v", err)
	}
	deleteID, err := st.UpsertBook(models.Book{FilePath: "/books-delete/sub/b.txt", Title: "delete", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert delete: %v", err)
	}
	if err := st.SaveProgress(models.ReadingProgress{BookID: deleteID, Chapter: 1, Page: 2}); err != nil {
		t.Fatalf("save progress: %v", err)
	}
	if _, err := st.AddBookmark(models.Bookmark{BookID: deleteID, Chapter: 0, Page: 0, Label: "mark"}); err != nil {
		t.Fatalf("bookmark: %v", err)
	}

	if err := st.DeleteBooksUnderDir("/books-delete"); err != nil {
		t.Fatalf("delete under dir: %v", err)
	}

	books, err := st.ListBooks()
	if err != nil {
		t.Fatalf("list books: %v", err)
	}
	if len(books) != 1 || books[0].ID != keepID {
		t.Fatalf("books: %+v", books)
	}
	if _, err := st.GetBook(deleteID); err == nil {
		t.Fatal("deleted book still exists")
	}
	progress, err := st.GetProgress(deleteID)
	if err != nil {
		t.Fatalf("get progress: %v", err)
	}
	if progress.Chapter != 0 || progress.Page != 0 {
		t.Fatalf("progress not cascaded: %+v", progress)
	}
	marks, err := st.ListBookmarks(deleteID)
	if err != nil {
		t.Fatalf("list bookmarks: %v", err)
	}
	if len(marks) != 0 {
		t.Fatalf("bookmarks not cascaded: %+v", marks)
	}
}

func TestDeleteBooksUnderDirKeepsSiblingDirectories(t *testing.T) {
	st := newTestStore(t)

	deleteID, err := st.UpsertBook(models.Book{FilePath: "/library/books/sub/a.txt", Title: "delete", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert delete: %v", err)
	}
	keepID, err := st.UpsertBook(models.Book{FilePath: "/library/bookshelf/b.txt", Title: "keep", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert keep: %v", err)
	}

	if err := st.DeleteBooksUnderDir("/library/books"); err != nil {
		t.Fatalf("delete under dir: %v", err)
	}

	if _, err := st.GetBook(deleteID); err == nil {
		t.Fatal("book under deleted dir still exists")
	}
	if got, err := st.GetBook(keepID); err != nil || got.ID != keepID {
		t.Fatalf("sibling book deleted: book=%+v err=%v", got, err)
	}
}

func TestDeleteBooksUnderDirTreatsPercentAndUnderscoreLiterally(t *testing.T) {
	st := newTestStore(t)

	deleteID, err := st.UpsertBook(models.Book{FilePath: "/library/books_%/sub/a.txt", Title: "delete", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert delete: %v", err)
	}
	underscoreSiblingID, err := st.UpsertBook(models.Book{FilePath: "/library/booksA%/sub/b.txt", Title: "keep underscore", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert underscore sibling: %v", err)
	}
	percentSiblingID, err := st.UpsertBook(models.Book{FilePath: "/library/books_long_name/sub/c.txt", Title: "keep percent", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert percent sibling: %v", err)
	}

	if err := st.DeleteBooksUnderDir("/library/books_%"); err != nil {
		t.Fatalf("delete under dir: %v", err)
	}

	if _, err := st.GetBook(deleteID); err == nil {
		t.Fatal("book under deleted wildcard dir still exists")
	}
	if got, err := st.GetBook(underscoreSiblingID); err != nil || got.ID != underscoreSiblingID {
		t.Fatalf("underscore wildcard sibling deleted: book=%+v err=%v", got, err)
	}
	if got, err := st.GetBook(percentSiblingID); err != nil || got.ID != percentSiblingID {
		t.Fatalf("percent wildcard sibling deleted: book=%+v err=%v", got, err)
	}
}

func TestDeleteBooksUnderDirDeletesNonASCIIChildPath(t *testing.T) {
	st := newTestStore(t)

	deleteID, err := st.UpsertBook(models.Book{FilePath: "/tmp/小说/a.txt", Title: "delete", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert delete: %v", err)
	}
	keepID, err := st.UpsertBook(models.Book{FilePath: "/tmp/小说家/a.txt", Title: "keep", Format: "txt"})
	if err != nil {
		t.Fatalf("upsert keep: %v", err)
	}

	if err := st.DeleteBooksUnderDir("/tmp/小说"); err != nil {
		t.Fatalf("delete under non-ASCII dir: %v", err)
	}

	if _, err := st.GetBook(deleteID); err == nil {
		t.Fatal("book under non-ASCII deleted dir still exists")
	}
	if got, err := st.GetBook(keepID); err != nil || got.ID != keepID {
		t.Fatalf("non-ASCII sibling book deleted: book=%+v err=%v", got, err)
	}
}
