package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/xuanchong/cli-read/models"
	"github.com/xuanchong/cli-read/ui"
)

type bookItem struct {
	book models.Book
}

func (b bookItem) Title() string       { return b.book.Title }
func (b bookItem) Description() string { return fmt.Sprintf("[%s] %s", strings.ToUpper(b.book.Format), b.book.Author) }
func (b bookItem) FilterValue() string { return b.book.Title }

// BookshelfModel is the bookshelf view.
type BookshelfModel struct {
	list      list.Model
	input     textinput.Model
	keys      ui.KeyMap
	searching bool
	status    string
	allItems  []list.Item
}

// NewBookshelfModel creates a new bookshelf model.
func NewBookshelfModel(books []models.Book) BookshelfModel {
	items := make([]list.Item, len(books))
	for i, b := range books {
		items[i] = bookItem{book: b}
	}
	l := list.New(items, list.NewDefaultDelegate(), 60, 20)
	l.Title = "书架"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "搜索书名..."
	ti.CharLimit = 50

	return BookshelfModel{
		list:     l,
		input:    ti,
		keys:     ui.DefaultKey(),
		allItems: items,
	}
}

// SetStatus sets the status bar text.
func (m *BookshelfModel) SetStatus(s string) {
	m.status = s
}

func (m BookshelfModel) Init() tea.Cmd { return nil }

func (m BookshelfModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searching = false
				m.input.Blur()
				m.input.SetValue("")
			case "enter":
				term := m.input.Value()
				m.searching = false
				m.input.Blur()
				m.input.SetValue("")
				m.filter(term)
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-4)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Search):
			m.searching = true
			m.input.Focus()
		case key.Matches(msg, m.keys.Open):
			if item, ok := m.list.SelectedItem().(bookItem); ok {
				return m, func() tea.Msg { return OpenBookMsg{Book: item.book} }
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *BookshelfModel) filter(term string) {
	term = strings.ToLower(term)
	if term == "" {
		m.list.SetItems(m.allItems)
		return
	}
	var items []list.Item
	for _, it := range m.allItems {
		bi := it.(bookItem)
		if strings.Contains(strings.ToLower(bi.book.Title), term) {
			items = append(items, bi)
		}
	}
	m.list.SetItems(items)
}

func (m BookshelfModel) View() string {
	var b strings.Builder
	b.WriteString(m.list.View())
	b.WriteString("\n")
	if m.searching {
		b.WriteString("/ ")
		b.WriteString(m.input.View())
	} else if m.status != "" {
		b.WriteString(ui.StatusStyle.Render(m.status))
	}
	return b.String()
}

// Selected returns the currently selected book (for testing).
func (m BookshelfModel) Selected() *models.Book {
	if item, ok := m.list.SelectedItem().(bookItem); ok {
		b := item.book
		return &b
	}
	return nil
}
