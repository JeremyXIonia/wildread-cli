package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/xuanchong/cli-read/config"
	"github.com/xuanchong/cli-read/models"
	"github.com/xuanchong/cli-read/ui"
)

type dirManagerMode int

const (
	dirModeList dirManagerMode = iota
	dirModeAdd
	dirModeConfirmCreate
	dirModeConfirmDelete
)

type DirectoryManagerModel struct {
	dirs        []models.LibraryDir
	selected    int
	mode        dirManagerMode
	input       textinput.Model
	keys        ui.KeyMap
	status      string
	pendingPath string
}

func NewDirectoryManagerModel(dirs []models.LibraryDir) DirectoryManagerModel {
	input := textinput.New()
	input.Placeholder = "粘贴或输入目录路径"
	input.CharLimit = 500
	return DirectoryManagerModel{
		dirs:  dirs,
		input: input,
		keys:  ui.DefaultKey(),
	}
}

func (m DirectoryManagerModel) Init() tea.Cmd { return nil }

func (m DirectoryManagerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.mode {
		case dirModeAdd:
			return m.updateAdd(msg)
		case dirModeConfirmCreate:
			return m.updateConfirmCreate(msg)
		case dirModeConfirmDelete:
			return m.updateConfirmDelete(msg)
		default:
			return m.updateList(msg)
		}
	}
	return m, nil
}

func (m DirectoryManagerModel) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return CloseDirectoryManagerMsg{} }
	case key.Matches(msg, m.keys.Down):
		if m.selected+1 < len(m.dirs) {
			m.selected++
		}
	case key.Matches(msg, m.keys.Up):
		if m.selected > 0 {
			m.selected--
		}
	case msg.String() == "a":
		m.mode = dirModeAdd
		m.status = ""
		m.pendingPath = ""
		m.input.SetValue("")
		m.input.Focus()
	case key.Matches(msg, m.keys.Delete):
		if len(m.dirs) > 0 {
			m.mode = dirModeConfirmDelete
		}
	case msg.String() == "r":
		return m, func() tea.Msg { return RescanLibraryDirsMsg{} }
	}
	return m, nil
}

func (m DirectoryManagerModel) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		m.mode = dirModeList
		m.pendingPath = ""
		m.input.Blur()
		return m, nil
	case key.Matches(msg, m.keys.Confirm):
		path, err := config.NormalizePath(m.input.Value())
		if err != nil {
			m.status = "目录不能为空"
			return m, nil
		}

		m.input.Blur()
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				m.mode = dirModeConfirmCreate
				m.pendingPath = path
				return m, nil
			}
			m.mode = dirModeList
			m.status = "无法检查目录：" + err.Error()
			return m, nil
		}
		if !info.IsDir() {
			m.mode = dirModeList
			m.status = "不是目录: " + path
			m.pendingPath = ""
			return m, nil
		}

		m.mode = dirModeList
		m.pendingPath = ""
		return m, func() tea.Msg { return AddLibraryDirMsg{Path: path, Create: false} }
	default:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
}

func (m DirectoryManagerModel) updateConfirmCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		path := m.pendingPath
		m.mode = dirModeList
		m.pendingPath = ""
		return m, func() tea.Msg { return AddLibraryDirMsg{Path: path, Create: true} }
	case "esc", "q", "n":
		m.mode = dirModeList
		m.pendingPath = ""
		return m, nil
	}
	return m, nil
}

func (m DirectoryManagerModel) updateConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		if len(m.dirs) == 0 || m.selected >= len(m.dirs) {
			m.mode = dirModeList
			return m, nil
		}
		dir := m.dirs[m.selected]
		m.mode = dirModeList
		return m, func() tea.Msg { return DeleteLibraryDirMsg{Dir: dir} }
	case "esc", "q", "n":
		m.mode = dirModeList
		return m, nil
	}
	return m, nil
}

func (m DirectoryManagerModel) View() string {
	switch m.mode {
	case dirModeAdd:
		var b strings.Builder
		b.WriteString("添加目录：\n")
		b.WriteString(m.input.View())
		b.WriteString("\n")
		if m.status != "" {
			b.WriteString("\n")
			b.WriteString(ui.StatusStyle.Render(m.status))
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(ui.HintStyle.Render("Enter 保存  Esc/q 取消"))
		return b.String()
	case dirModeConfirmCreate:
		return fmt.Sprintf("目录不存在：%s\n\n是否创建该目录并加入书籍目录？\n\n输入 y 确认创建，Esc/q/n 取消", m.pendingPath)
	case dirModeConfirmDelete:
		if len(m.dirs) == 0 || m.selected >= len(m.dirs) {
			return "没有可删除的目录\n\n" + ui.HintStyle.Render("Esc/q 返回")
		}
		dir := m.dirs[m.selected]
		return fmt.Sprintf("删除目录：%s\n\n这会删除该目录下已入库的书籍、阅读进度和书签。\n目录中的原始文件不会被删除。\n\n输入 y 确认删除，Esc/q 取消", dir.Path)
	default:
		var b strings.Builder
		b.WriteString("书籍目录\n\n")
		if len(m.dirs) == 0 {
			b.WriteString(ui.HintStyle.Render("（暂无目录）"))
			b.WriteString("\n")
		}
		for i, dir := range m.dirs {
			marker := "  "
			if i == m.selected {
				marker = "> "
			}
			b.WriteString(marker)
			b.WriteString(dir.Path)
			if dir.IsDefault {
				b.WriteString("        默认")
			}
			b.WriteString("\n")
		}
		if m.status != "" {
			b.WriteString("\n")
			b.WriteString(ui.StatusStyle.Render(m.status))
		}
		b.WriteString("\n")
		b.WriteString(ui.HintStyle.Render("a 添加目录  d 删除目录  r 重新扫描  q 返回书架"))
		return b.String()
	}
}
