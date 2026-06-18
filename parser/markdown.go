package parser

import "github.com/xuanchong/cli-read/models"

// ParseMarkdown parses a Markdown file (placeholder).
func ParseMarkdown(path string) (*models.Book, error) {
	return &models.Book{Format: "md"}, nil
}
