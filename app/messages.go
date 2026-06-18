package app

import "github.com/xuanchong/cli-read/models"

// OpenBookMsg is sent when the user selects a book to read.
type OpenBookMsg struct {
	Book models.Book
}

// OpenDirectoryManagerMsg is sent when the user opens book directory management.
type OpenDirectoryManagerMsg struct{}

// CloseDirectoryManagerMsg is sent when the user exits directory management.
type CloseDirectoryManagerMsg struct{}

// AddLibraryDirMsg asks the root model to add a managed book directory.
type AddLibraryDirMsg struct {
	Path   string
	Create bool
}

// DeleteLibraryDirMsg asks the root model to delete a managed directory and its book records.
type DeleteLibraryDirMsg struct {
	Dir models.LibraryDir
}

// RescanLibraryDirsMsg asks the root model to rescan all managed directories.
type RescanLibraryDirsMsg struct{}
