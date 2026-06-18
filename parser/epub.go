package parser

import "github.com/xuanchong/cli-read/models"

// ParseEPUB parses an EPUB file (placeholder).
func ParseEPUB(path string) (*models.Book, error) {
	return &models.Book{Format: "epub"}, nil
}
