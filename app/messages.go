package app

import "github.com/xuanchong/cli-read/models"

// OpenBookMsg is sent when the user selects a book to read.
type OpenBookMsg struct {
	Book models.Book
}
