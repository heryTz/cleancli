package main

import (
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

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
	total    int
}

type unit struct {
	suffix string
	base   float64
}

var units = []unit{
	{suffix: "EB", base: math.Pow10(18)},
	{suffix: "PB", base: math.Pow10(15)},
	{suffix: "TB", base: math.Pow10(12)},
	{suffix: "GB", base: math.Pow10(9)},
	{suffix: "MB", base: math.Pow10(6)},
	{suffix: "KB", base: math.Pow10(3)},
	{suffix: "B", base: math.Pow10(0)},
}

func humanByte(size int) string {
	val := float64(size)
	suffix := "B"
	for _, unit := range units {
		res := float64(size) / unit.base
		if res >= 1 {
			val = res
			suffix = unit.suffix
			break
		}
	}
	return fmt.Sprintf("%.1f %s", val, suffix)
}

func getCacheDir(dir string) string {
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		return path.Join(home, dir[2:])
	}
	return dir
}

var CACHE_DIR = getCacheDir("./cache-test")

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

func getTotalSize(files []FileModel) int {
	total := 0
	for _, file := range files {
		total += file.size
	}
	return total
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

func initialModel() model {
	files, err := scanDir(CACHE_DIR)
	if err != nil {
		panic(err)
	}
	total := 0
	selected := map[int]struct{}{
		0: {},
	}
	for i, file := range files {
		total += file.size
		selected[i+1] = struct{}{}
	}
	return model{
		files:    files,
		selected: selected,
		total:    total,
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

			total := 0
			for i, file := range m.files {
				if _, ok := m.selected[i+1]; ok {
					total += file.size
				}
			}
			m.total = total

		case "enter":
			for i, file := range m.files {
				if _, ok := m.selected[i+1]; ok {
					if file.isDir {
						os.RemoveAll(path.Join(CACHE_DIR, file.name))
					} else {
						os.Remove(path.Join(CACHE_DIR, file.name))
					}

				}
			}
			files, err := scanDir(CACHE_DIR)
			if err != nil {
				panic(err)
			}
			m.files = files
			m.selected = make(map[int]struct{})
			m.cursor = 0
		}
	}

	return m, nil
}

func (m model) View() string {
	s := "Select cache that you want to delete?\n\n"

	if len(m.files) == 0 {
		s += "Empty cache\n"
	} else {
		s += fmt.Sprintf("Total: %s\n\n", humanByte(m.total))

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
	s += "\n- Press Space to select item."
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
