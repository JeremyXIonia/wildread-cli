package parser

import "testing"

func TestHTMLToTextParagraphs(t *testing.T) {
	html := `<html><body><p>第一段。</p><p>第二段。</p></body></html>`
	title, paras := HTMLToText(html)
	if title != "" {
		t.Errorf("title: %q", title)
	}
	if len(paras) != 2 || paras[0] != "第一段。" || paras[1] != "第二段。" {
		t.Fatalf("paragraphs: %+v", paras)
	}
}

func TestHTMLToTextHeading(t *testing.T) {
	html := `<html><body><h1>第一章</h1><p>正文。</p></body></html>`
	title, paras := HTMLToText(html)
	if title != "第一章" {
		t.Errorf("title: %q", title)
	}
	if len(paras) != 1 || paras[0] != "正文。" {
		t.Fatalf("paragraphs: %+v", paras)
	}
}

func TestHTMLToTextStripTags(t *testing.T) {
	html := `<p>这是 <b>加粗</b> 和 <i>斜体</i> 的文本。</p>`
	_, paras := HTMLToText(html)
	if len(paras) != 1 || paras[0] != "这是 加粗 和 斜体 的文本。" {
		t.Fatalf("paragraphs: %+v", paras)
	}
}

func TestHTMLToTextEmpty(t *testing.T) {
	title, paras := HTMLToText("")
	if title != "" || len(paras) != 0 {
		t.Fatalf("empty: %q %v", title, paras)
	}
}
