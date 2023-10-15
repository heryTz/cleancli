package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type MyFile struct {
	Info         fs.FileInfo
	Checked      bool
	Children     []MyFile
	RelativePath string
}

type model struct {
	files    []MyFile
	cursor   int
	selected map[int]struct{}
}

func initialModel() model {
	files, err := scanDirectory("./cache-test")
	if err != nil {
		panic(err)
	}
	return model{
		files:    files,
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "down":
			if m.cursor == len(m.files)-1 {
				m.cursor = 0
			} else {
				m.cursor++
			}
		case "up":
			if m.cursor == 0 {
				m.cursor = len(m.files) - 1
			} else {
				m.cursor--
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		case "enter":
			println("Delete cache")
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Select cache that you want to delete?\n\n"

	if len(m.files) == 0 {
		s += "Empty cache\n"
	} else {
		cursor := " "
		if m.cursor == 0 {
			cursor = ">"
		}

		allChecked := " "
		if len(m.selected) == len(m.files) {
			allChecked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s", cursor, allChecked, "Select all\n\n")

		for i, file := range m.files {
			cursor := " "
			if i == m.cursor+1 && m.cursor != 0 {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.selected[i]; ok {
				checked = "x"
			}

			s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, file.Info.Name())
		}
	}

	s += "\n- Press Enter to delete cache."
	s += "\n- Press q to quit."
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func scanDirectory(dir string) ([]MyFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []MyFile{}
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}
		relativePath := filepath.Join(dir, info.Name())
		if info.IsDir() {
			children, err := scanDirectory(relativePath)
			if err != nil {
				return nil, err
			}
			files = append(files, MyFile{Info: info, RelativePath: relativePath, Children: children})
		} else {
			files = append(files, MyFile{Info: info, RelativePath: relativePath})
		}
	}

	return files, nil
}

func main2() {
	myFiles, err := scanDirectory("./cache-test")
	if err != nil {
		panic(err)
	}

	debugFile(myFiles, 0)
}

func debugFile(files []MyFile, depth int) {
	for _, file := range files {
		fmt.Printf("%s%s\n", strings.Repeat("\t", depth), file.RelativePath)
		if file.Children != nil {
			debugFile(file.Children, depth+1)
		}
	}
}
