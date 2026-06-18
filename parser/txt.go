package parser

import "github.com/xuanchong/cli-read/models"

// ParseTXT parses a TXT file (placeholder).
func ParseTXT(path string) (*models.Book, error) {
	return &models.Book{Format: "txt"}, nil
}
