package main

import (
	"fmt"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
)

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

func getDirSize(dir string) (int, error) {
	root := dir
	fileSystem := os.DirFS(root)
	size := 0
	err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !d.IsDir() {
			size += int(info.Size())
		}
		return nil
	})

	return size, err
}

func getTotalSize(files []fileModel) int {
	total := 0
	for _, file := range files {
		total += file.size
	}
	return total
}

func scanDir(dir string) ([]fileModel, error) {
	files := []fileModel{}
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
				// some file is not readable
				// log.Print(err)
			}
			if size > 0 {
				files = append(files, fileModel{name: info.Name(), size: size, isDir: true})
			}
		} else {
			files = append(files, fileModel{name: info.Name(), size: int(info.Size()), isDir: false})
		}
	}

	return files, nil
}
