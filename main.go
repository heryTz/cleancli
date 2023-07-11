package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type MyFile struct {
	Info         fs.FileInfo
	Checked      bool
	Children     []MyFile
	RelativePath string
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

func main() {
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
