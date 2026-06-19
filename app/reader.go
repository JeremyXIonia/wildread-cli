package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/JeremyXIonia/wildread-cli/models"
	"github.com/JeremyXIonia/wildread-cli/pager"
	"github.com/JeremyXIonia/wildread-cli/store"
	"github.com/JeremyXIonia/wildread-cli/ui"
)

// ReaderMode represents the current sub-mode of the reader.
type ReaderMode int

const (
	ModeReading ReaderMode = iota
	ModeTOC
	ModeBookmarks
)

// ReaderModel is the reading view.
type ReaderModel struct {
	book             *models.Book
	pager            *pager.Pager
	viewport         viewport.Model
	keys             ui.KeyMap
	chapter          int
	page             int
	width            int
	height           int
	mode             ReaderMode
	status           string
	store            *store.Store
	bookmarks        []models.Bookmark
	tocSelected      int
	bookmarkSelected int
}

const (
	readerHeaderLines = 1
	readerFooterLines = 1
)

// NewReaderModel creates a new reader model.
func NewReaderModel(book *models.Book, progress models.ReadingProgress, st *store.Store) ReaderModel {
	chapter := progress.Chapter
	page := progress.Page
	if len(book.Chapters) == 0 {
		book.Chapters = []models.Chapter{{Title: "空", Content: "（无内容）"}}
	}
	if chapter < 0 || chapter >= len(book.Chapters) {
		chapter = 0
	}
	if page < 0 {
		page = 0
	}

	p := pager.New(book.Chapters[chapter].Content, 80, 20)

	vp := viewport.New(80, 18)
	// 禁用 viewport 默认按键，由 reader 自己处理翻页
	vp.KeyMap = viewport.KeyMap{}

	var bms []models.Bookmark
	if st != nil {
		bms, _ = st.ListBookmarks(book.ID)
	}

	return ReaderModel{
		book:      book,
		pager:     p,
		viewport:  vp,
		keys:      ui.DefaultKey(),
		chapter:   chapter,
		page:      page,
		mode:      ModeReading,
		store:     st,
		bookmarks: bms,
	}
}

func (m ReaderModel) Init() tea.Cmd {
	return nil
}

// IsReading reports whether the reader is showing the main reading view.
func (m ReaderModel) IsReading() bool {
	return m.mode == ModeReading
}

func (m *ReaderModel) loadPageContent() {
	if m.page >= m.pager.PageCount() {
		m.page = m.pager.PageCount() - 1
	}
	if m.page < 0 {
		m.page = 0
	}
	content, err := m.pager.Page(m.page)
	if err != nil {
		m.status = "页码错误"
		return
	}
	m.viewport.SetContent(content)
	m.status = ""
}

func (m ReaderModel) saveProgress() tea.Cmd {
	return func() tea.Msg {
		if m.store == nil || m.book == nil {
			return nil
		}
		_ = m.store.SaveProgress(models.ReadingProgress{
			BookID:  m.book.ID,
			Chapter: m.chapter,
			Page:    m.page,
		})
		return nil
	}
}

func (m ReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case ModeReading:
			return m.updateReading(msg)
		case ModeTOC:
			return m.updateTOC(msg)
		case ModeBookmarks:
			return m.updateBookmarks(msg)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m ReaderModel) updateReading(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, m.saveProgress()
	case key.Matches(msg, m.keys.Down):
		m.nextPage()
	case key.Matches(msg, m.keys.Up):
		m.prevPage()
	case key.Matches(msg, m.keys.GotoTop):
		m.page = 0
		m.loadPageContent()
	case key.Matches(msg, m.keys.GotoEnd):
		m.page = m.pager.PageCount() - 1
		m.loadPageContent()
	case key.Matches(msg, m.keys.Next):
		m.nextChapter()
	case key.Matches(msg, m.keys.Prev):
		m.prevChapter()
	case key.Matches(msg, m.keys.Open):
		m.tocSelected = m.chapter
		m.mode = ModeTOC
	case key.Matches(msg, m.keys.Bookmarks):
		if m.store != nil {
			m.bookmarks, _ = m.store.ListBookmarks(m.book.ID)
		}
		m.bookmarkSelected = 0
		m.mode = ModeBookmarks
	case key.Matches(msg, m.keys.Mark):
		if m.store == nil {
			m.status = "无数据库"
			break
		}
		id, err := m.store.AddBookmark(models.Bookmark{
			BookID:  m.book.ID,
			Chapter: m.chapter,
			Page:    m.page,
			Label:   fmt.Sprintf("ch%d p%d", m.chapter+1, m.page+1),
		})
		if err != nil {
			m.status = "加书签失败"
		} else {
			m.status = fmt.Sprintf("已加书签 #%d", id)
		}
	}
	return m, m.saveProgress()
}

func (m ReaderModel) updateTOC(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = ModeReading
	case key.Matches(msg, m.keys.Down):
		if m.tocSelected+1 < len(m.book.Chapters) {
			m.tocSelected++
		}
	case key.Matches(msg, m.keys.Up):
		if m.tocSelected > 0 {
			m.tocSelected--
		}
	case key.Matches(msg, m.keys.Confirm):
		m.chapter = m.tocSelected
		m.page = 0
		m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
		m.loadPageContent()
		m.mode = ModeReading
		return m, m.saveProgress()
	}
	return m, nil
}

func (m ReaderModel) updateBookmarks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = ModeReading
	case key.Matches(msg, m.keys.Down):
		if m.bookmarkSelected+1 < len(m.bookmarks) {
			m.bookmarkSelected++
		}
	case key.Matches(msg, m.keys.Up):
		if m.bookmarkSelected > 0 {
			m.bookmarkSelected--
		}
	case key.Matches(msg, m.keys.Confirm):
		if len(m.bookmarks) == 0 {
			return m, nil
		}
		bm := m.bookmarks[m.bookmarkSelected]
		if bm.Chapter >= 0 && bm.Chapter < len(m.book.Chapters) {
			m.chapter = bm.Chapter
			m.page = bm.Page
			m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
			m.loadPageContent()
		}
		m.mode = ModeReading
		return m, m.saveProgress()
	}
	return m, nil
}

func (m *ReaderModel) nextPage() {
	if m.page+1 < m.pager.PageCount() {
		m.page++
	} else if m.chapter+1 < len(m.book.Chapters) {
		m.chapter++
		m.page = 0
		m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
	}
	m.loadPageContent()
}

func (m *ReaderModel) prevPage() {
	if m.page > 0 {
		m.page--
	} else if m.chapter > 0 {
		m.chapter--
		m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
		m.page = m.pager.PageCount() - 1
	}
	m.loadPageContent()
}

func (m *ReaderModel) nextChapter() {
	if m.chapter+1 < len(m.book.Chapters) {
		m.chapter++
		m.page = 0
		m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
	}
	m.loadPageContent()
}

func (m *ReaderModel) prevChapter() {
	if m.chapter > 0 {
		m.chapter--
		m.page = 0
		m.pager = pager.New(m.book.Chapters[m.chapter].Content, m.viewport.Width, m.viewport.Height)
	}
	m.loadPageContent()
}

func (m *ReaderModel) resize() {
	bodyHeight := m.height - readerHeaderLines - readerFooterLines
	if bodyHeight < 5 {
		bodyHeight = 5
	}
	bodyWidth := m.width
	if bodyWidth < 20 {
		bodyWidth = 20
	}
	m.viewport.Width = bodyWidth
	m.viewport.Height = bodyHeight
	m.pager = pager.New(m.book.Chapters[m.chapter].Content, bodyWidth, bodyHeight)
	m.loadPageContent()
}

func (m ReaderModel) header() string {
	chTitle := m.book.Chapters[m.chapter].Title
	if chTitle == "" {
		chTitle = fmt.Sprintf("第 %d 章", m.chapter+1)
	}
	return fmt.Sprintf("《%s》 - %s", m.book.Title, chTitle)
}

func (m ReaderModel) footer() string {
	total := m.pager.PageCount()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("第 %d/%d 页 | 第 %d/%d 章", m.page+1, total, m.chapter+1, len(m.book.Chapters)))
	if m.status != "" {
		b.WriteString("  ")
		b.WriteString(ui.StatusStyle.Render(m.status))
	}
	b.WriteString("  ")
	b.WriteString(ui.HintStyle.Render("j/k翻 b签 m签 o录 n/p章 q返"))
	return b.String()
}

func (m ReaderModel) View() string {
	if m.mode == ModeTOC {
		return m.viewTOC()
	}
	if m.mode == ModeBookmarks {
		return m.viewBookmarks()
	}
	var b strings.Builder
	b.WriteString(ui.TitleStyle.Render(m.header()))
	b.WriteString("\n")
	b.WriteString(m.viewport.View())
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

func (m ReaderModel) listBodyHeight() int {
	bodyHeight := m.height - 4
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	return bodyHeight
}

func visibleListRange(total, selected, height int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}
	if height >= total {
		return 0, total
	}
	start := selected - height/2
	if start < 0 {
		start = 0
	}
	if start+height > total {
		start = total - height
	}
	return start, start + height
}

func (m ReaderModel) viewTOC() string {
	var b strings.Builder
	b.WriteString("章节目录\n\n")
	start, end := visibleListRange(len(m.book.Chapters), m.tocSelected, m.listBodyHeight())
	for i := start; i < end; i++ {
		c := m.book.Chapters[i]
		marker := "  "
		if i == m.tocSelected {
			marker = "> "
		}
		title := c.Title
		if title == "" {
			title = fmt.Sprintf("第 %d 章", i+1)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", marker, title))
	}
	b.WriteString("\n")
	b.WriteString(ui.HintStyle.Render("j/k 移动  Enter 跳转  esc/q 返回阅读"))
	return b.String()
}

func (m ReaderModel) viewBookmarks() string {
	var b strings.Builder
	b.WriteString("书签\n\n")
	if len(m.bookmarks) == 0 {
		b.WriteString(ui.HintStyle.Render("（暂无书签）"))
	} else {
		start, end := visibleListRange(len(m.bookmarks), m.bookmarkSelected, m.listBodyHeight())
		for i := start; i < end; i++ {
			bm := m.bookmarks[i]
			marker := "  "
			if i == m.bookmarkSelected {
				marker = "> "
			}
			b.WriteString(fmt.Sprintf("%s#%d  第 %d 章 第 %d 页  %s\n", marker, bm.ID, bm.Chapter+1, bm.Page+1, bm.Label))
		}
	}
	b.WriteString("\n")
	b.WriteString(ui.HintStyle.Render("j/k 移动  Enter 跳转  esc/q 返回阅读"))
	return b.String()
}
