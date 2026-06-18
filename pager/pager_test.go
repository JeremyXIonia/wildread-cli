package pager

import "testing"

func TestPagerEnglish(t *testing.T) {
	text := "line one..\nline two..\nline three\nline four.\nline five."
	p := New(text, 10, 3)
	if got := p.PageCount(); got != 2 {
		t.Fatalf("page count: %d", got)
	}
	p1, _ := p.Page(0)
	want1 := "line one..\nline two..\nline three"
	if p1 != want1 {
		t.Fatalf("page 0: %q", p1)
	}
	p2, _ := p.Page(1)
	want2 := "line four.\nline five."
	if p2 != want2 {
		t.Fatalf("page 1: %q", p2)
	}
}

func TestPagerChinese(t *testing.T) {
	text := "你好世界\n你好世界\n你好世界"
	p := New(text, 8, 2)
	if got := p.PageCount(); got != 2 {
		t.Fatalf("page count: %d", got)
	}
}

func TestPagerEmpty(t *testing.T) {
	p := New("", 10, 5)
	if p.PageCount() != 1 {
		t.Fatalf("empty count: %d", p.PageCount())
	}
	c, _ := p.Page(0)
	if c != "" {
		t.Fatalf("empty: %q", c)
	}
}

func TestPagerOutOfRange(t *testing.T) {
	p := New("a\nb\nc", 10, 2)
	if _, err := p.Page(-1); err == nil {
		t.Fatal("expected error for -1")
	}
	if _, err := p.Page(99); err == nil {
		t.Fatal("expected error for 99")
	}
}
