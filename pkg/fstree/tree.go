package fstree

/*
base.BookmarkNodeBase based on StackOverflow answer by user 'icza' from: https://stackoverflow.com/a/32962550
*/

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/util"
	"gopkg.in/djherbis/times.v1"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//PopulateTmpDir reflects abstract trees to the filesystem as files and dirs
func PopulateTmpDir(baseDirs map[string]*base.BookmarkNodeBase, tracker util.BookmarkTracker) (string, map[string]string) {
	programName := "bookmark2fs"
	tmpDirPath, err := ioutil.TempDir(os.TempDir(), programName)
	check(err)

	populateRootDir := func(rootNode *base.BookmarkNodeBase) string {
		stack := []*base.BookmarkNodeBase{rootNode}
		filename := util.StringToFilename(rootNode.Name)
		rootNode.Path = path.Join(tmpDirPath, filename)
		bmIndex := rootNode.Name + rootNode.URL
		tracker.In[bmIndex] = rootNode.Path
		for len(stack) > 0 {
			fileNode := stack[len(stack)-1]
			nodePath := fileNode.Path
			stack = stack[:len(stack)-1]

			// initialBMMap[rootNode.Name+rootNode.URL] = rootNode.Path

			switch fileNode.Type {
			case "url":
				d := []byte(fileNode.URL + "\n")

				err := ioutil.WriteFile(fileNode.Path, d, 0644)
				check(err)
				// set file times according to date added
				ctime := time.Unix(fileNode.DateCreated, 0)
				atime := time.Unix(fileNode.DateModified, 0)
				err = os.Chtimes(fileNode.Path, atime, ctime)
				check(err)

			case "folder":
				err := os.Mkdir(nodePath, 0755)
				check(err)

				// defer os.RemoveAll(nodePath)

				for _, child := range fileNode.Children {
					childFName := util.StringToFilename(child.Name)
					child.Path = path.Join(nodePath, childFName)
					stack = append(stack, child)
				}
			default:
				panic("Unrecognized node type:" + string(fileNode.Type))
			}
		}
		// defer os.RemoveAll(tmpRoot)

		return rootNode.Path
	}

	tmpRootPaths := make(map[string]string)
	for k, v := range baseDirs {
		rootPath := populateRootDir(v)
		tmpRootPaths[k] = rootPath
	}

	// TODO open fs handler of os
	return tmpDirPath, tmpRootPaths
}

//ConstructFSTree construct file trees according to
func ConstructFSTree(path string, tracker util.BookmarkTracker) *base.BookmarkNodeBase {

	// currently only handles URL
	toFileObj := func(file os.FileInfo, t times.Timespec, path string) *base.BookmarkNodeBase {
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
		JSONFile := base.BookmarkNodeBase{
			DateModified: int64(t.AccessTime().Unix()),
			DateCreated:  int64(t.ModTime().Unix()),
			Type:         nodeType,
			Name:         util.FilenameToString(file.Name()),
			URL:          URL,
			Path:         path,
			Children:     []*base.BookmarkNodeBase{},
		}
		tracker.Out[JSONFile.Name+URL] = path
		// if file.Mode()&os.ModeSymlink == os.ModeSymlink {
		// 	// JSONFile.IsLink = true
		// 	JSONFile.LinksTo, _ = filepath.EvalSymlinks(filepath.Join(path, file.Name()))
		// } // Else case is the zero values of the fields
		return &JSONFile
	}
	rootOSFile, _ := os.Stat(path)
	t, err := times.Stat(path)
	check(err)

	rootFile := toFileObj(rootOSFile, t, path) //start with root file
	stack := []*base.BookmarkNodeBase{rootFile}

	for len(stack) > 0 { //until stack is empty,
		file := stack[len(stack)-1] //pop entry from stack
		t, err := times.Stat(file.Path)
		check(err)
		stack = stack[:len(stack)-1]
		children, _ := ioutil.ReadDir(file.Path) //get the children of entry
		for _, chld := range children {          //for each child
			child := toFileObj(chld, t, filepath.Join(file.Path, chld.Name())) //turn it into a base.BookmarkNodeBase object
			file.Children = append(file.Children, child)                       //append it to the children of the current file popped
			stack = append(stack, child)                                       //append the child to the stack, so the same process can be run again
		}
	}

	return rootFile
	// output, _ := json.MarshalIndent(rootFile, "", "     ")
	// fmt.Println(string(output))
}
