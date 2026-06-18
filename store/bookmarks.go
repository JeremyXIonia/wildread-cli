package store

import "github.com/xuanchong/cli-read/models"

func (s *Store) AddBookmark(b models.Bookmark) (int64, error) {
    res, err := s.db.Exec(
        `INSERT INTO bookmarks (book_id, chapter, page, label) VALUES (?, ?, ?, ?)`,
        b.BookID, b.Chapter, b.Page, b.Label)
    if err != nil { return 0, err }
    return res.LastInsertId()
}

func (s *Store) ListBookmarks(bookID int64) ([]models.Bookmark, error) {
    rows, err := s.db.Query(
        `SELECT id, book_id, chapter, page, COALESCE(label, ''), created_at
         FROM bookmarks WHERE book_id=? ORDER BY id`, bookID)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []models.Bookmark
    for rows.Next() {
        var b models.Bookmark
        if err := rows.Scan(&b.ID, &b.BookID, &b.Chapter, &b.Page, &b.Label, &b.CreatedAt); err != nil {
            return nil, err
        }
        out = append(out, b)
    }
    return out, rows.Err()
}

func (s *Store) DeleteBookmark(id int64) error {
    _, err := s.db.Exec(`DELETE FROM bookmarks WHERE id=?`, id)
    return err
}
