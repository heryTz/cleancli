package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

type FileModel struct {
	name  string
	size  int
	isDir bool
}

type model struct {
	files    []FileModel
	cursor   int
	selected map[int]struct{}
}

func initialModel() model {
	files, err := scanDir("./cache-test")
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
		// select all + len m.files
		cursorLen := len(m.files)
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "down":
			if m.cursor == cursorLen {
				m.cursor = 0
			} else {
				m.cursor++
			}
		case "up":
			if m.cursor == 0 {
				m.cursor = cursorLen
			} else {
				m.cursor--
			}
		case " ":
			_, ok := m.selected[m.cursor]
			if m.cursor == 0 && ok {
				m.selected = make(map[int]struct{})
			} else if m.cursor == 0 && !ok {
				m.selected = map[int]struct{}{
					0: {},
				}
				for i := range m.files {
					m.selected[i+1] = struct{}{}
				}
			} else if ok {
				delete(m.selected, m.cursor)
				delete(m.selected, 0)
			} else {
				m.selected[m.cursor] = struct{}{}
				if len(m.selected) == cursorLen {
					m.selected[0] = struct{}{}
				}
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
		checkAllCursor := " "
		if m.cursor == 0 {
			checkAllCursor = ">"
		}

		allChecked := " "
		if _, ok := m.selected[0]; ok {
			allChecked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s", checkAllCursor, allChecked, "Select all\n\n")

		for i, file := range m.files {
			cursor := " "
			fileCursor := i + 1
			if fileCursor == m.cursor && m.cursor != 0 {
				cursor = ">"
			}

			checked := " "
			if _, ok := m.selected[fileCursor]; ok {
				checked = "x"
			}

			s += fmt.Sprintf("%s [%s] %s (%s)\n", cursor, checked, file.name, humanByte(file.size))
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

func humanByte(size int) string {
	return fmt.Sprintf("%d", size)
}

func getDirSize(dir string) (int, error) {
	root := dir
	fileSystem := os.DirFS(root)
	size := 0
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}
		info, err := d.Info()
		if err != nil {
			panic(err)
		}
		if !d.IsDir() {
			size += int(info.Size())
		}
		return nil
	})

	return size, err
}

func scanDir(dir string) ([]FileModel, error) {
	files := []FileModel{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			panic(err)
		}

		if entry.IsDir() {
			size, err := getDirSize(filepath.Join(dir, info.Name()))
			if err != nil {
				panic(err)
			}
			files = append(files, FileModel{name: info.Name(), size: size, isDir: true})
		} else {
			files = append(files, FileModel{name: info.Name(), size: int(info.Size()), isDir: false})
		}
	}

	return files, nil
}
