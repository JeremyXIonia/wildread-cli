package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings used across the application.
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	GotoTop    key.Binding
	GotoEnd    key.Binding
	Open       key.Binding
	Search     key.Binding
	ManageDirs key.Binding
	Back       key.Binding
	Quit       key.Binding
	Next       key.Binding
	Prev       key.Binding
	Mark       key.Binding
	Bookmarks  key.Binding
	Confirm    key.Binding
	Delete     key.Binding
}

// DefaultKey returns the default key bindings.
func DefaultKey() KeyMap {
	return KeyMap{
		Up:         key.NewBinding(key.WithKeys("k", "up", "pgup"), key.WithHelp("k/pgup/↑", "上一页")),
		Down:       key.NewBinding(key.WithKeys("j", "down", "pgdown", " "), key.WithHelp("j/space/pgdn/↓", "下一页")),
		GotoTop:    key.NewBinding(key.WithKeys("g", "home"), key.WithHelp("g", "到首")),
		GotoEnd:    key.NewBinding(key.WithKeys("G", "end"), key.WithHelp("G", "到尾")),
		Open:       key.NewBinding(key.WithKeys("o", "enter"), key.WithHelp("o/enter", "打开")),
		Search:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "搜索")),
		ManageDirs: key.NewBinding(key.WithKeys("d", "D"), key.WithHelp("d", "目录")),
		Back:       key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc/q", "返回")),
		Quit:       key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "退出")),
		Next:       key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "下一章")),
		Prev:       key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "上一章")),
		Mark:       key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "加书签")),
		Bookmarks:  key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "书签")),
		Confirm:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "确认")),
		Delete:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "删除")),
	}
}
