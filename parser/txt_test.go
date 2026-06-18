package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseTXTBasic(t *testing.T) {
	book, err := ParseTXT(filepath.Join("..", "testdata", "sample.txt"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if book.Title == "" {
		t.Fatal("title empty")
	}
	if len(book.Chapters) != 1 {
		t.Fatalf("chapters: %d", len(book.Chapters))
	}
	c := book.Chapters[0]
	if c.Title != "" {
		t.Errorf("chapter title: %q", c.Title)
	}
	if !strings.Contains(c.Content, "第一段内容。") ||
		!strings.Contains(c.Content, "第二段内容。") ||
		!strings.Contains(c.Content, "第三段内容。") {
		t.Fatalf("content: %q", c.Content)
	}
}
