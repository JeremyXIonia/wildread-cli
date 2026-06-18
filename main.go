package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/xuanchong/cli-read/app"
	"github.com/xuanchong/cli-read/config"
	"github.com/xuanchong/cli-read/models"
	"github.com/xuanchong/cli-read/parser"
	"github.com/xuanchong/cli-read/store"
)

type rootModel struct {
	dataDir        string
	defaultBookDir string
	tempBookDir    string
	store          *store.Store
	mode           appMode
	bookshelf      app.BookshelfModel
	reader         *app.ReaderModel
}

type appMode int

const (
	modeBookshelf appMode = iota
	modeReader
)

func main() {
	dataDirFlag := flag.String("data-dir", "", "应用数据目录")
	tempDirFlag := flag.String("dir", "", "临时书籍目录（本次扫描，不保存）")
	dbPathFlag := flag.String("db", "", "SQLite 数据库路径")
	flag.Parse()

	paths, err := config.ResolvePaths(*dataDirFlag, *dbPathFlag, *tempDirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "解析路径失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(paths.DataDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "无法创建应用数据目录: %v\n", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(paths.DefaultBookDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "无法创建默认书籍目录: %v\n", err)
		os.Exit(1)
	}

	st, err := store.Open(paths.DBPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开数据库: %v\n", err)
		os.Exit(1)
	}
	defer st.Close()

	defaultCreated, err := ensureDefaultLibraryDir(st, paths.DefaultBookDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化默认书籍目录失败: %v\n", err)
		os.Exit(1)
	}

	scanDirs, err := configuredScanDirs(st, paths.TempBookDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取书籍目录失败: %v\n", err)
		os.Exit(1)
	}

	books, scanErrs, err := refreshBooks(st, scanDirs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "同步书架失败: %v\n", err)
		os.Exit(1)
	}

	root := rootModel{
		dataDir:        paths.DataDir,
		defaultBookDir: paths.DefaultBookDir,
		tempBookDir:    paths.TempBookDir,
		store:          st,
		mode:           modeBookshelf,
		bookshelf:      app.NewBookshelfModel(books),
	}
	root.bookshelf.SetStatus(startupStatus(len(books), defaultCreated, scanErrs, paths.DefaultBookDir, paths.TempBookDir))

	p := tea.NewProgram(root, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "运行错误: %v\n", err)
		os.Exit(1)
	}
}

func ensureDefaultLibraryDir(st *store.Store, defaultDir string) (bool, error) {
	dirs, err := st.ListLibraryDirs()
	if err != nil {
		return false, err
	}
	if len(dirs) > 0 {
		return false, nil
	}
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		return false, err
	}
	_, err = st.AddLibraryDir(defaultDir, true)
	return err == nil, err
}

func configuredScanDirs(st *store.Store, tempDir string) ([]string, error) {
	libraryDirs, err := st.ListLibraryDirs()
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(libraryDirs)+1)
	for _, d := range libraryDirs {
		dirs = append(dirs, d.Path)
	}
	if tempDir != "" {
		dirs = append(dirs, tempDir)
	}
	return dirs, nil
}

type scanResult struct {
	paths           []string
	successfulRoots []string
	failedRoots     []string
	errs            []error
}

func scanAllDirs(dirs []string) ([]string, []error) {
	result := scanDirs(dirs)
	return result.paths, result.errs
}

func scanDirs(dirs []string) scanResult {
	seen := map[string]bool{}
	var result scanResult
	for _, dir := range dirs {
		root := absolutePath(dir)
		scanned, err := parser.Scan(dir)
		if err != nil {
			result.failedRoots = append(result.failedRoots, root)
			result.errs = append(result.errs, fmt.Errorf("%s: %w", dir, err))
			continue
		}
		result.successfulRoots = append(result.successfulRoots, root)
		for _, p := range scanned {
			abs := absolutePath(p)
			if !seen[abs] {
				seen[abs] = true
				result.paths = append(result.paths, abs)
			}
		}
	}
	return result
}

func refreshBooks(st *store.Store, dirs []string) ([]models.Book, []error, error) {
	result := scanDirs(dirs)
	if err := syncBooksForRoots(st, result.paths, result.successfulRoots, result.failedRoots); err != nil {
		return nil, result.errs, err
	}
	books, err := st.ListBooks()
	return books, result.errs, err
}

func startupStatus(bookCount int, defaultCreated bool, scanErrs []error, defaultBookDir, tempBookDir string) string {
	if len(scanErrs) > 0 {
		return fmt.Sprintf("已扫描 %d 本书，%d 个目录扫描失败", bookCount, len(scanErrs))
	}
	if defaultCreated {
		return fmt.Sprintf("未配置书籍目录，已使用默认目录 %s", defaultBookDir)
	}
	if tempBookDir != "" {
		return fmt.Sprintf("已临时扫描目录 %s；如需长期使用，请在目录管理中添加", tempBookDir)
	}
	return fmt.Sprintf("已扫描 %d 本书", bookCount)
}

func syncBooks(st *store.Store, paths []string) error {
	return syncBooksForRoots(st, paths, nil, nil)
}

func syncBooksForRoots(st *store.Store, paths []string, pruneRoots, preserveRoots []string) error {
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
		abs := absolutePath(p)
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
		if !seen[b.FilePath] && shouldPruneBook(b.FilePath, pruneRoots, preserveRoots) {
			if err := st.DeleteBook(b.ID); err != nil {
				fmt.Fprintf(os.Stderr, "删除失败 %d: %v\n", b.ID, err)
			}
		}
	}
	return nil
}

func shouldPruneBook(bookPath string, pruneRoots, preserveRoots []string) bool {
	absBookPath := absolutePath(bookPath)
	for _, root := range preserveRoots {
		if pathWithinRoot(absBookPath, root) {
			return false
		}
	}
	if len(pruneRoots) == 0 {
		return true
	}
	for _, root := range pruneRoots {
		if pathWithinRoot(absBookPath, root) {
			return true
		}
	}
	return false
}

func pathWithinRoot(path, root string) bool {
	if path == root {
		return true
	}
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && rel != ".." && !startsWithDotDot(rel)
}

func startsWithDotDot(path string) bool {
	return len(path) > 3 && path[:3] == ".."+string(os.PathSeparator)
}

func absolutePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	return abs
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case app.OpenBookMsg:
		// 重新解析文件以获取章节内容（store 只存了 metadata）
		parsed, err := parser.ParseByExtension(msg.Book.FilePath)
		if err != nil {
			m.bookshelf.SetStatus(fmt.Sprintf("解析失败: %v", err))
			return m, nil
		}
		parsed.ID = msg.Book.ID
		parsed.FilePath = msg.Book.FilePath
		progress, _ := m.store.GetProgress(msg.Book.ID)
		reader := app.NewReaderModel(parsed, progress, m.store)
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
		if m.mode == modeReader && m.reader != nil && m.reader.IsReading() && (msg.String() == "esc" || msg.String() == "q") {
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
