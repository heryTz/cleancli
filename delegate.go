package main

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	byteSizeStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, l *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(fileModel)
	if !ok {
		return
	}

	checked := " "
	if i.checked {
		checked = "x"
	}

	size := byteSizeStyle.Render("(" + humanByte(i.size) + ")")
	str := fmt.Sprintf("[%s] %d. %s %s", checked, index+1, i.name, size)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = selectedItemStyle.Render
	}

	fmt.Fprint(w, fn(str))
}
