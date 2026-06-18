package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xuanchong/cli-read/app"
	"github.com/xuanchong/cli-read/parser"
	"github.com/xuanchong/cli-read/store"
)

type rootModel struct {
	dir       string
	store     *store.Store
	mode      appMode
	bookshelf app.BookshelfModel
	reader    *app.ReaderModel
}

type appMode int

const (
	modeBookshelf appMode = iota
	modeReader
)

func main() {
	dir := flag.String("dir", "./books", "书籍目录")
	dbPath := flag.String("db", "./novel-reader.db", "SQLite 数据库路径")
	flag.Parse()

	if err := os.MkdirAll(*dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "无法创建书籍目录: %v\n", err)
		os.Exit(1)
	}

	st, err := store.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开数据库: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	paths, err := parser.Scan(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "扫描目录失败: %v\n", err)
		os.Exit(1)
	}

	if err := syncBooks(st, paths); err != nil {
		fmt.Fprintf(os.Stderr, "同步书架失败: %v\n", err)
		os.Exit(1)
	}

	books, err := st.ListBooks()
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取书架失败: %v\n", err)
		os.Exit(1)
	}

	root := rootModel{
		dir:       *dir,
		store:     st,
		mode:      modeBookshelf,
		bookshelf: app.NewBookshelfModel(books),
	}
	root.bookshelf.SetStatus(fmt.Sprintf("已扫描 %d 本书", len(books)))

	p := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "运行错误: %v\n", err)
		os.Exit(1)
	}
}

func syncBooks(st *store.Store, paths []string) error {
	existing, err := st.ListBooks()
	if err != nil {
		return err
	}

	existingByPath := map[string]int64{}
	for _, b := range existing {
		existingByPath[b.FilePath] = b.ID
	}

	seen := map[string]bool{}
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			abs = p
		}
		seen[abs] = true
		if _, ok := existingByPath[abs]; ok {
			continue
		}
		book, err := parser.ParseByExtension(abs)
		if err != nil {
			fmt.Fprintf(os.Stderr, "解析失败 %s: %v\n", abs, err)
			continue
		}
		book.FilePath = abs
		if _, err := st.UpsertBook(*book); err != nil {
			fmt.Fprintf(os.Stderr, "入库失败 %s: %v\n", abs, err)
		}
	}

	for _, b := range existing {
		if !seen[b.FilePath] {
			if err := st.DeleteBook(b.ID); err != nil {
				fmt.Fprintf(os.Stderr, "删除失败 %d: %v\n", b.ID, err)
			}
		}
	}
	return nil
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case app.OpenBookMsg:
		progress, _ := m.store.GetProgress(msg.Book.ID)
		reader := app.NewReaderModel(&msg.Book, progress, m.store)
		m.reader = &reader
		m.mode = modeReader
		return m, nil

	case tea.WindowSizeMsg:
		if m.mode == modeBookshelf {
			bs, cmd := m.bookshelf.Update(msg)
			m.bookshelf = bs.(app.BookshelfModel)
			return m, cmd
		}

	case tea.KeyMsg:
		if m.mode == modeReader && (msg.String() == "esc" || msg.String() == "q") {
			m.mode = modeBookshelf
			return m, nil
		}
	}

	if m.mode == modeBookshelf {
		bs, cmd := m.bookshelf.Update(msg)
		m.bookshelf = bs.(app.BookshelfModel)
		return m, cmd
	}
	if m.mode == modeReader && m.reader != nil {
		var rm app.ReaderModel
		nm, cmd := m.reader.Update(msg)
		rm = nm.(app.ReaderModel)
		m.reader = &rm
		return m, cmd
	}
	return m, nil
}

func (m rootModel) View() string {
	if m.mode == modeReader && m.reader != nil {
		return m.reader.View()
	}
	return m.bookshelf.View()
}
