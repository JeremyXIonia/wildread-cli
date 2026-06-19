package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/JeremyXIonia/wildread-cli/models"
	"golang.org/x/text/transform"
)

// ParseTXT parses a TXT file, auto-detects encoding, and extracts content.
func ParseTXT(path string) (*models.Book, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read txt: %w", err)
	}

	raw = stripBOM(raw)

	text, err := decodeBytes(raw)
	if err != nil {
		return nil, fmt.Errorf("decode txt: %w", err)
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	blocks := strings.Split(text, "\n\n")
	var paragraphs []string
	for _, b := range blocks {
		b = strings.TrimSpace(b)
		if b != "" {
			paragraphs = append(paragraphs, b)
		}
	}
	content := strings.Join(paragraphs, "\n\n")

	base := filepath.Base(path)
	title := strings.TrimSuffix(base, filepath.Ext(base))

	return &models.Book{
		Title:    title,
		Author:   "",
		Format:   "txt",
		Chapters: []models.Chapter{{Title: "", Content: content}},
	}, nil
}

// stripBOM removes UTF-8 BOM bytes.
func stripBOM(b []byte) []byte {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return b[3:]
	}
	return b
}

// decodeBytes attempts to decode bytes as UTF-8, GB18030, or GBK.
func decodeBytes(b []byte) (string, error) {
	if utf8.Valid(b) {
		return string(b), nil
	}
	for _, label := range []string{"gb18030", "gbk"} {
		enc, err := lookupEncoding(label)
		if err != nil {
			continue
		}
		reader := transform.NewReader(newBytesReader(b), enc.NewDecoder())
		decoded, err := readAll(reader)
		if err == nil {
			return decoded, nil
		}
	}
	return "", fmt.Errorf("unable to detect encoding")
}
