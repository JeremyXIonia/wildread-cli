package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestParseEPUB(t *testing.T) {
	book, err := ParseEPUB(filepath.Join("..", "testdata", "sample.epub"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if book.Title != "测试书" {
		t.Errorf("title: %q", book.Title)
	}
	if book.Author != "测试作者" {
		t.Errorf("author: %q", book.Author)
	}
	if len(book.Chapters) != 2 {
		t.Fatalf("chapters: %d", len(book.Chapters))
	}
	if book.Chapters[0].Title != "第一章 开始" {
		t.Errorf("ch0 title: %q", book.Chapters[0].Title)
	}
	if !strings.Contains(book.Chapters[0].Content, "第一段内容") {
		t.Errorf("ch0 content: %q", book.Chapters[0].Content)
	}
	if book.Chapters[1].Title != "第二章 继续" {
		t.Errorf("ch1 title: %q", book.Chapters[1].Title)
	}
}
