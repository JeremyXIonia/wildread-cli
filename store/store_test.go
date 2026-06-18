package store

import (
    "os"
    "path/filepath"
    "testing"
    "github.com/xuanchong/cli-read/models"
)

func newTestStore(t *testing.T) *Store {
    t.Helper()
    path := filepath.Join(t.TempDir(), "test.db")
    s, err := Open(path)
    if err != nil { t.Fatalf("open: %v", err) }
    t.Cleanup(func() { s.Close(); os.Remove(path) })
    return s
}

func TestOpenClose(t *testing.T) {
    s := newTestStore(t)
    if err := s.Close(); err != nil { t.Fatalf("close: %v", err) }
}

func TestBooksCRUD(t *testing.T) {
    s := newTestStore(t)
    b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
    id, err := s.UpsertBook(b)
    if err != nil || id <= 0 { t.Fatalf("upsert: %v, %d", err, id) }
    got, _ := s.GetBook(id)
    if got.Title != "A" || got.Format != "epub" { t.Fatalf("bad book: %+v", got) }
    list, _ := s.ListBooks()
    if len(list) != 1 { t.Fatalf("list: %d", len(list)) }
    s.DeleteBook(id)
    list, _ = s.ListBooks()
    if len(list) != 0 { t.Fatalf("not empty") }
}

func TestUpsertBookUpdates(t *testing.T) {
    s := newTestStore(t)
    b := models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"}
    id, _ := s.UpsertBook(b)
    b.ID = id; b.Title = "A-new"
    id2, _ := s.UpsertBook(b)
    if id != id2 { t.Fatalf("id mismatch") }
    got, _ := s.GetBook(id)
    if got.Title != "A-new" { t.Fatalf("not updated") }
}

func TestProgressCRUD(t *testing.T) {
    s := newTestStore(t)
    id, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
    p, _ := s.GetProgress(id)
    if p.Chapter != 0 || p.Page != 0 { t.Fatalf("default: %+v", p) }
    s.SaveProgress(models.ReadingProgress{BookID: id, Chapter: 2, Page: 5})
    p, _ = s.GetProgress(id)
    if p.Chapter != 2 || p.Page != 5 { t.Fatalf("saved: %+v", p) }
}

func TestBookmarksCRUD(t *testing.T) {
    s := newTestStore(t)
    bid, _ := s.UpsertBook(models.Book{FilePath: "/a.epub", Title: "A", Format: "epub"})
    id1, _ := s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 1, Page: 2, Label: "m1"})
    s.AddBookmark(models.Bookmark{BookID: bid, Chapter: 3, Page: 4, Label: "m2"})
    list, _ := s.ListBookmarks(bid)
    if len(list) != 2 || list[0].Label != "m1" { t.Fatalf("list: %+v", list) }
    s.DeleteBookmark(id1)
    list, _ = s.ListBookmarks(bid)
    if len(list) != 1 { t.Fatalf("after delete: %d", len(list)) }
}
