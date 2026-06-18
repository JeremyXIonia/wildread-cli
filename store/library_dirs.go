package store

import (
	"path/filepath"
	"strings"

	"github.com/xuanchong/cli-read/models"
)

func (s *Store) ListLibraryDirs() ([]models.LibraryDir, error) {
	rows, err := s.db.Query(`SELECT id, path, is_default, created_at FROM library_dirs ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.LibraryDir
	for rows.Next() {
		var d models.LibraryDir
		var isDefault int
		if err := rows.Scan(&d.ID, &d.Path, &isDefault, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.IsDefault = isDefault != 0
		out = append(out, d)
	}
	return out, rows.Err()
}

func (s *Store) AddLibraryDir(path string, isDefault bool) (int64, error) {
	defaultInt := 0
	if isDefault {
		defaultInt = 1
	}
	res, err := s.db.Exec(`INSERT INTO library_dirs (path, is_default) VALUES (?, ?)`, path, defaultInt)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteLibraryDir(id int64) error {
	_, err := s.db.Exec(`DELETE FROM library_dirs WHERE id=?`, id)
	return err
}

func (s *Store) DeleteBooksUnderDir(dir string) error {
	dir = filepath.Clean(dir)
	prefix := dir + string(filepath.Separator) + "%"
	if strings.HasSuffix(dir, string(filepath.Separator)) {
		prefix = dir + "%"
	}
	_, err := s.db.Exec(`DELETE FROM books WHERE file_path = ? OR file_path LIKE ?`, dir, prefix)
	return err
}
