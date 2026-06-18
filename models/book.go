package models

type Book struct {
	ID       int64
	FilePath string
	Title    string
	Author   string
	Format   string
	Chapters []Chapter
}

type Chapter struct {
	Title   string
	Content string
}

type ReadingProgress struct {
	BookID  int64
	Chapter int
	Page    int
}

type Bookmark struct {
	ID        int64
	BookID    int64
	Chapter   int
	Page      int
	Label     string
	CreatedAt string
}
