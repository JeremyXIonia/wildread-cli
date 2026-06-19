package store

import (
	"database/sql"
	"errors"
	"github.com/JeremyXIonia/wildread-cli/models"
)

func (s *Store) UpsertBook(b models.Book) (int64, error) {
	if b.ID == 0 {
		res, err := s.db.Exec(
			`INSERT INTO books (file_path, title, author, format) VALUES (?, ?, ?, ?)`,
			b.FilePath, b.Title, b.Author, b.Format)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	}
	_, err := s.db.Exec(
		`UPDATE books SET title=?, author=?, format=? WHERE id=?`,
		b.Title, b.Author, b.Format, b.ID)
	return b.ID, err
}

func (s *Store) GetBook(id int64) (models.Book, error) {
	var b models.Book
	err := s.db.QueryRow(
		`SELECT id, file_path, title, COALESCE(author, ''), format FROM books WHERE id=?`,
		id).Scan(&b.ID, &b.FilePath, &b.Title, &b.Author, &b.Format)
	if errors.Is(err, sql.ErrNoRows) {
		return b, errors.New("not found")
	}
	return b, err
}

func (s *Store) ListBooks() ([]models.Book, error) {
	rows, err := s.db.Query(
		`SELECT id, file_path, title, COALESCE(author, ''), format FROM books ORDER BY added_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Book
	for rows.Next() {
		var b models.Book
		if err := rows.Scan(&b.ID, &b.FilePath, &b.Title, &b.Author, &b.Format); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) DeleteBook(id int64) error {
	_, err := s.db.Exec(`DELETE FROM books WHERE id=?`, id)
	return err
}
