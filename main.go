package main

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle      = lipgloss.NewStyle()
	paginationStyle = list.DefaultStyles().PaginationStyle.PaddingLeft(2)
	helpStyle       = list.DefaultStyles().HelpStyle.PaddingBottom(1)
	CACHE_DIR       = getCacheDir("~/Library/Caches")
)

const (
	listHeight = 19
	listWidth  = 20
)

type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Confirm key.Binding
	Select  key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Confirm, k.Select},
		{k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down"),
		key.WithHelp("↓", "move down"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm delete"),
	),
	Select: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle selection"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
}

type model struct {
	total          int
	totalSelected  int
	keys           keyMap
	list           list.Model
	loading        bool
	scanDirLoading bool
	scanDirSpinner spinner.Model
}

type unit struct {
	suffix string
	base   float64
}

type finishDelete int
type loadedDir []fileModel

type fileModel struct {
	name    string
	size    int
	isDir   bool
	checked bool
}

func (i fileModel) FilterValue() string {
	return i.name
}

var p *tea.Program

func scanDirAsync(dir string) {
	files, err := scanDir(dir)
	if err != nil {
		log.Fatalf("failed to scan directory %s", dir)
	}
	p.Send(loadedDir(files))
}

func (m model) Init() tea.Cmd {
	go scanDirAsync(CACHE_DIR)
	return m.scanDirSpinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case loadedDir:
		files := msg
		items := []list.Item{}
		for _, file := range files {
			items = append(items, fileModel{
				name:  file.name,
				size:  file.size,
				isDir: file.isDir,
			})
		}
		m.scanDirLoading = false
		m.total = getTotalSize(files)
		m.list.SetItems(items)

	case finishDelete:
		files := []list.Item{}
		total := 0
		for _, file := range m.list.Items() {
			f := file.(fileModel)
			if !f.checked {
				files = append(files, f)
				total += f.size
			}
			// wtf! why m.list.RemoveItem() not work properly
		}
		m.loading = false
		m.total = total
		m.totalSelected = 0
		m.list.StopSpinner()
		m.list.SetItems(files)

	case tea.KeyMsg:
		if m.loading {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Select):
			selected, _ := m.list.SelectedItem().(fileModel)
			for i, file := range m.list.Items() {
				f := file.(fileModel)
				if selected.name == f.name {
					f.checked = !f.checked
					m.list.SetItem(i, f)
					if f.checked {
						m.totalSelected += f.size
					} else {
						m.totalSelected -= f.size
					}
				}
			}
		case key.Matches(msg, m.keys.Confirm):
			m.loading = true

			go func() {
				for _, file := range m.list.Items() {
					f := file.(fileModel)
					if f.checked {
						if f.isDir {
							os.RemoveAll(path.Join(CACHE_DIR, f.name))
						} else {
							os.Remove(path.Join(CACHE_DIR, f.name))
						}
					}
				}
				p.Send(finishDelete(0))
			}()

			return m, m.list.StartSpinner()
		}
	}

	var cmd []tea.Cmd
	var cmdList tea.Cmd
	m.list, cmdList = m.list.Update(msg)
	cmd = append(cmd, cmdList)

	if m.scanDirLoading {
		var cmdSpin tea.Cmd
		m.scanDirSpinner, cmdSpin = m.scanDirSpinner.Update(msg)
		cmd = append(cmd, cmdSpin)
	}

	return m, tea.Batch(cmd...)
}

func (m model) View() string {
	if m.scanDirLoading {
		return fmt.Sprintf("\n%s Scan cache...", m.scanDirSpinner.View())
	}

	totalPaddingLeft := ""
	if m.loading {
		totalPaddingLeft = "  "
	}

	m.list.Title = fmt.Sprintf("%s\n\n%sTotal: %s / %s", m.list.Title, totalPaddingLeft, humanByte(m.totalSelected), humanByte(m.total))
	return "\n" + m.list.View()
}

func main() {
	list := list.New([]list.Item{}, itemDelegate{}, listWidth, listHeight)
	list.Title = "Select cache that you want to delete?"
	list.Styles.Title = titleStyle
	list.Styles.PaginationStyle = paginationStyle
	list.Styles.HelpStyle = helpStyle
	list.SetFilteringEnabled(false)

	spinner := spinner.New()

	p = tea.NewProgram(model{
		keys:           keys,
		list:           list,
		scanDirLoading: true,
		scanDirSpinner: spinner,
	})

	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
