package parser

import (
	"strings"

	"golang.org/x/net/html"
)

// HTMLToText extracts plain text from HTML.
// The first <h1>~<h6> is returned as the title (and removed from body).
// <p>/<div> are treated as paragraph separators. <br> becomes newline.
// All other tags are stripped.
func HTMLToText(input string) (string, []string) {
	doc, err := html.Parse(strings.NewReader(input))
	if err != nil {
		return "", nil
	}

	var title string
	var paragraphs []string
	var current strings.Builder

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			current.WriteString(n.Data)
			return
		}
		if n.Type != html.ElementNode && n.Type != html.DocumentNode {
			return
		}
		tag := strings.ToLower(n.Data)

		if tag == "script" || tag == "style" || tag == "title" || tag == "head" {
			return
		}

		if tag == "br" {
			current.WriteString("\n")
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}

		isHeading := tag == "h1" || tag == "h2" || tag == "h3" || tag == "h4" || tag == "h5" || tag == "h6"
		if isHeading && title == "" {
			t := strings.TrimSpace(current.String())
			if t != "" {
				title = t
				current.Reset()
				return
			}
		}

		if tag == "p" || tag == "div" {
			text := strings.TrimSpace(current.String())
			if text != "" {
				paragraphs = append(paragraphs, text)
			}
			current.Reset()
		}
	}

	walk(doc)

	if t := strings.TrimSpace(current.String()); t != "" && title == "" {
		title = t
	}

	return title, paragraphs
}
