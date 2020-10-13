/*
bookmarkNodeBase based on StackOverflow answer by user 'icza' from: https://stackoverflow.com/a/32962550
*/
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/djherbis/times.v1"
)

// construct  according to
func constructFSTree(path string) *bookmarkNodeBase {
	rootOSFile, _ := os.Stat(path)
	t, err := times.Stat(path)
	check(err)

	rootFile := toFile(rootOSFile, t, path) //start with root file
	stack := []*bookmarkNodeBase{rootFile}

	for len(stack) > 0 { //until stack is empty,
		file := stack[len(stack)-1] //pop entry from stack
		t, err := times.Stat(file.Path)
		check(err)
		stack = stack[:len(stack)-1]
		children, _ := ioutil.ReadDir(file.Path) //get the children of entry
		for _, chld := range children {          //for each child
			child := toFile(chld, t, filepath.Join(file.Path, chld.Name())) //turn it into a bookmarkNodeBase object
			file.Children = append(file.Children, child)                    //append it to the children of the current file popped
			stack = append(stack, child)                                    //append the child to the stack, so the same process can be run again
		}
	}

	return rootFile
	// output, _ := json.MarshalIndent(rootFile, "", "     ")
	// fmt.Println(string(output))
}

// currently only handles URL
func toFile(file os.FileInfo, t times.Timespec, path string) *bookmarkNodeBase {
	URL := ""
	nodeType := "folder"
	if !file.IsDir() {
		nodeType = "url" // TODO could use enums here instead

		// fmt.Println(path)
		content, err := ioutil.ReadFile(path)
		check(err)

		contents := strings.SplitN(string(content), "\n", -1)
		//only first line contains non-whitespace chars
		if len(contents) > 2 || contents[1] != "" {
			fmt.Println("Warning, only first line containing URL in file '" + path + "' was parsed.")
		}
		URL = contents[0]
	}
	JSONFile := bookmarkNodeBase{
		DateModified: int64(t.AccessTime().Unix()),
		DateCreated:  int64(t.ModTime().Unix()),
		Type:         nodeType,
		Name:         filenameToString(file.Name()),
		URL:          URL,
		Path:         path,
		Children:     []*bookmarkNodeBase{},
	}
	// if file.Mode()&os.ModeSymlink == os.ModeSymlink {
	// 	// JSONFile.IsLink = true
	// 	JSONFile.LinksTo, _ = filepath.EvalSymlinks(filepath.Join(path, file.Name()))
	// } // Else case is the zero values of the fields
	return &JSONFile
}
