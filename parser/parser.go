package parser

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xuanchong/cli-read/models"
)

// FormatFromExt returns the format identifier ("epub"/"txt"/"md"/"").
func FormatFromExt(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".epub":
		return "epub"
	case ".txt":
		return "txt"
	case ".md", ".markdown":
		return "md"
	default:
		return ""
	}
}

// ParseByExtension selects a parser based on file extension.
func ParseByExtension(path string) (*models.Book, error) {
	switch FormatFromExt(path) {
	case "epub":
		return ParseEPUB(path)
	case "txt":
		return ParseTXT(path)
	case "md":
		return ParseMarkdown(path)
	default:
		return nil, fmt.Errorf("unsupported format: %s", path)
	}
}
