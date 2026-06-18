package store

import (
    "database/sql"
    "errors"
    "github.com/xuanchong/cli-read/models"
)

func (s *Store) GetProgress(bookID int64) (models.ReadingProgress, error) {
    var p models.ReadingProgress
    p.BookID = bookID
    err := s.db.QueryRow(
        `SELECT chapter, page FROM reading_progress WHERE book_id=?`,
        bookID).Scan(&p.Chapter, &p.Page)
    if errors.Is(err, sql.ErrNoRows) { return p, nil }
    return p, err
}

func (s *Store) SaveProgress(p models.ReadingProgress) error {
    _, err := s.db.Exec(`
        INSERT INTO reading_progress (book_id, chapter, page, updated_at)
        VALUES (?, ?, ?, CURRENT_TIMESTAMP)
        ON CONFLICT(book_id) DO UPDATE SET
            chapter=excluded.chapter, page=excluded.page, updated_at=CURRENT_TIMESTAMP`,
        p.BookID, p.Chapter, p.Page)
    return err
}
