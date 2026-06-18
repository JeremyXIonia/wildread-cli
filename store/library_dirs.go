package store

import (
	"database/sql"
	"errors"
	"path/filepath"

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
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var dir string
	if err := tx.QueryRow(`SELECT path FROM library_dirs WHERE id=?`, id).Scan(&dir); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if err := deleteBooksUnderDir(tx, dir); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM library_dirs WHERE id=?`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) DeleteBooksUnderDir(dir string) error {
	return deleteBooksUnderDir(s.db, dir)
}

type booksUnderDirDeleter interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func deleteBooksUnderDir(db booksUnderDirDeleter, dir string) error {
	dir = filepath.Clean(dir)
	childPrefix := dir + string(filepath.Separator)
	_, err := db.Exec(
		`DELETE FROM books WHERE file_path = ? OR substr(file_path, 1, ?) = ?`,
		dir, len(childPrefix), childPrefix,
	)
	return err
}
