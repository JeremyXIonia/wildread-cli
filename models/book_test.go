package models

import "testing"

func TestBookFields(t *testing.T) {
	b := Book{ID: 1, FilePath: "/x.epub", Title: "X", Format: "epub"}
	if b.ID != 1 || b.Title != "X" || b.Format != "epub" {
		t.Fatalf("unexpected book: %+v", b)
	}
}

func TestChapterContent(t *testing.T) {
	c := Chapter{Title: "ch1", Content: "p1\n\np2"}
	if c.Title != "ch1" || c.Content != "p1\n\np2" {
		t.Fatalf("unexpected chapter: %+v", c)
	}
}
