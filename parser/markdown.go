package parser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JeremyXIonia/wildread-cli/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var md = goldmark.New()

// ParseMarkdown parses a Markdown file, splitting into chapters by #/## headings.
func ParseMarkdown(path string) (*models.Book, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read md: %w", err)
	}

	reader := text.NewReader(raw)
	root := md.Parser().Parse(reader)

	var chapters []models.Chapter
	var current models.Chapter
	flush := func() {
		if current.Title != "" || strings.TrimSpace(current.Content) != "" {
			current.Content = strings.TrimSpace(current.Content)
			chapters = append(chapters, current)
		}
		current = models.Chapter{}
	}

	err = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch v := n.(type) {
		case *ast.Heading:
			flush()
			level := v.Level
			txt := nodeText(v, raw)
			if level <= 2 {
				current.Title = strings.TrimSpace(txt)
			} else {
				current.Content += strings.Repeat("#", level) + " " + txt + "\n\n"
			}
		case *ast.Paragraph:
			current.Content += strings.TrimSpace(nodeText(v, raw)) + "\n\n"
		}
		return ast.WalkContinue, nil
	})
	if err != nil {
		return nil, err
	}
	flush()

	base := filepath.Base(path)
	title := strings.TrimSuffix(base, filepath.Ext(base))

	return &models.Book{
		Title:    title,
		Format:   "md",
		Chapters: chapters,
	}, nil
}

// nodeText extracts plain text from an AST node.
func nodeText(n ast.Node, source []byte) string {
	var buf bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		seg := n.Lines().At(i)
		buf.Write(seg.Value(source))
	}
	return buf.String()
}
