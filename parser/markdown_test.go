package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMarkdown(t *testing.T) {
	book, err := ParseMarkdown(filepath.Join("..", "testdata", "sample.md"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if book.Title == "" {
		t.Fatal("title empty")
	}
	if len(book.Chapters) != 2 {
		t.Fatalf("chapters: %d", len(book.Chapters))
	}
	if book.Chapters[0].Title != "第一章 开始" {
		t.Errorf("ch0 title: %q", book.Chapters[0].Title)
	}
	if !strings.Contains(book.Chapters[0].Content, "第一章的第一段") {
		t.Errorf("ch0 content: %q", book.Chapters[0].Content)
	}
	if book.Chapters[1].Title != "第二章 继续" {
		t.Errorf("ch1 title: %q", book.Chapters[1].Title)
	}
}
