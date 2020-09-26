/*
File based on StackOverflow answer by user 'icza' from: https://stackoverflow.com/a/32962550
*/
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

type File struct {
	ModifiedTime time.Time `json:"ModifiedTime"`
	IsLink       bool      `json:"IsLink"`
	Type         string    `json:"IsDir"` // todo make enum type
	LinksTo      string    `json:"LinksTo"`
	Size         int64     `json:"Size"`
	Name         string    `json:"Name"`
	Path         string    `json:"Path"`
	Children     []*File   `json:"Children"`
}

func iterateJSON(path string) {
	rootOSFile, _ := os.Stat(path)
	rootFile := toFile(rootOSFile, path) //start with root file
	stack := []*File{rootFile}

	for len(stack) > 0 { //until stack is empty,
		file := stack[len(stack)-1] //pop entry from stack
		stack = stack[:len(stack)-1]
		children, _ := ioutil.ReadDir(file.Path) //get the children of entry
		for _, chld := range children {          //for each child
			child := toFile(chld, filepath.Join(file.Path, chld.Name())) //turn it into a File object
			file.Children = append(file.Children, child)                 //append it to the children of the current file popped
			stack = append(stack, child)                                 //append the child to the stack, so the same process can be run again
		}
	}

	output, _ := json.MarshalIndent(rootFile, "", "     ")
	fmt.Println(string(output))
}

func toFile(file os.FileInfo, path string) *File {
	nodeType := "url" // could use enums here instead
	if file.IsDir() {
		nodeType = "dir"
	}
	JSONFile := File{ModifiedTime: file.ModTime(),
		Type:     nodeType,
		Size:     file.Size(),
		Name:     file.Name(),
		Path:     path,
		Children: []*File{},
	}
	if file.Mode()&os.ModeSymlink == os.ModeSymlink {
		JSONFile.IsLink = true
		JSONFile.LinksTo, _ = filepath.EvalSymlinks(filepath.Join(path, file.Name()))
	} // Else case is the zero values of the fields
	return &JSONFile
}
