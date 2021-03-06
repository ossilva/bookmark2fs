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

	"github.com/google/uuid"
	"github.com/ossilva/bookmark2fs/pkg/configuration"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/util"
	"gopkg.in/djherbis/times.v1"
)

type nodeHistoryStack struct {
	// would be better use pointer to top to avoid duplication
	stack   []*base.BookmarkNodeBase
	history []*base.BookmarkNodeBase
}

func newNodeHistoryStack(initialNode *base.BookmarkNodeBase) *nodeHistoryStack {
	nodeHistoryStack := new(nodeHistoryStack)
	nodeHistoryStack.history = []*base.BookmarkNodeBase{initialNode}
	nodeHistoryStack.stack = []*base.BookmarkNodeBase{initialNode}
	return nodeHistoryStack
}

func (historyStack *nodeHistoryStack) push(node *base.BookmarkNodeBase) {
	historyStack.stack = append(historyStack.stack, node)
	historyStack.history = append(historyStack.history, node)
}

func (historyStack *nodeHistoryStack) pop() *base.BookmarkNodeBase {
	stack := &historyStack.stack
	node := (*stack)[len(*stack)-1]
	*stack = (*stack)[:len(*stack)-1]
	return node
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func mkFolderNameFile(pathString string, folderPath string) {
	folderNamePath := pathString + "/~FOLDERNAME"
	err := ioutil.WriteFile(folderNamePath, []byte(filepath.Base(folderPath)+"\n"), 0644)
	if err != nil {
		fmt.Println(fmt.Sprintf("Could not write to temporary directory: %s", pathString))
	}
}

func getUniqFilename(filepath string) string {
	var tryPath = filepath
	count := 1
	dedupPath := func() {
		tryPath = fmt.Sprintf("%s%s", filepath, strings.Repeat("^", count))
		count++
	}

	for {
		info, err := os.Stat(tryPath)
		if err != nil {
			if os.IsNotExist(err) {
				break
			} else {
				panic(err)
			}
		} else if info.IsDir() {
			dedupPath()
		} else if !info.IsDir() {
			dedupPath()
		}
	}
	return tryPath
}

func mkFileTreeDir(tmpRoot string) string {
	if tmpRoot == "" {
		tmpDirPath, err := ioutil.TempDir(os.TempDir(), configuration.ProgramName)
		check(err)
		mkFolderNameFile(tmpDirPath, "TEMP_ROOT")
		return tmpDirPath
	}

	_, err := os.Stat(tmpRoot)
	if os.IsNotExist(err) {
		mkdirErr := os.MkdirAll(tmpRoot, 0755)
		check(mkdirErr)
		mkFolderNameFile(tmpRoot, "TEMP_ROOT")
		return tmpRoot
	}

	if err == nil {
		mkFolderNameFile(tmpRoot, "TEMP_ROOT")
		return tmpRoot
	}

	panic("Given temporary directory path already exists")

}
func createURLFile(node *base.BookmarkNodeBase, filename string) {
	contents := []byte(node.URL + "\n" + node.Name + "\n")
	err := ioutil.WriteFile(filename, contents, 0644)
	check(err)
}

func createFolderDir(node *base.BookmarkNodeBase, filename string) {
	mkdirErr := os.Mkdir(filename, 0755)
	check(mkdirErr)
	mkFolderNameFile(filename, node.Path)
}

func stackFolderChildren(stack *nodeHistoryStack, children []*base.BookmarkNodeBase, nodePath string) {
	for _, child := range children {
		childFName := util.StringToFilename(child.Name)
		child.Path = path.Join(nodePath, childFName)
		stack.push(child)
	}
}

//PopulateTmpDir reflects abstract trees to the filesystem as files and dirs
func PopulateTmpDir(
	baseDirs []*base.BookmarkNodeBase,
	tracker *util.BookmarkTracker,
	tmpRoot string,
) (string, []string) {
	tmpDirPath := mkFileTreeDir(tmpRoot)

	populateRootDir := func(rootNode *base.BookmarkNodeBase) string {
		//TODO creation time is changed in sequence after file creation
		historyStack := newNodeHistoryStack(rootNode)
		filename := util.StringToFilename(rootNode.Name)
		rootNode.Path = path.Join(tmpDirPath, filename)
		for len(historyStack.stack) > 0 {
			fileNode := historyStack.pop()
			tracker.Insert(fileNode)

			nodePath := fileNode.Path
			newFilename := getUniqFilename(nodePath)

			switch fileNode.Type {
			case "url":
				createURLFile(fileNode, newFilename)
			case "folder":
				createFolderDir(fileNode, newFilename)
				stackFolderChildren(historyStack, fileNode.Children, nodePath)
			default:
				panic("Unrecognized node type:" + string(fileNode.Type))
			}
		}
		// times set after file tree population due to inode linking
		for i := range historyStack.history {
			// set file times according to date added
			node := historyStack.history[len(historyStack.history)-1-i] // leaves added after branches
			ctime := time.Unix(node.DateCreated, 0)
			atime := time.Unix(node.DateModified, 0)
			err := os.Chtimes(node.Path, atime, ctime)
			check(err)
		}

		return rootNode.Path
	}

	tmpRootPaths := []string{}
	for _, v := range baseDirs {
		rootPath := populateRootDir(v)
		tmpRootPaths = append(tmpRootPaths, rootPath)
	}

	// TODO open fs handler of os
	return tmpDirPath, tmpRootPaths
}

func getFolderName(path string) string {
	folderName, err := ioutil.ReadFile(path + "/~FOLDERNAME")
	check(err)

	contents := strings.SplitN(string(folderName), "\n", -1)

	//only first line contains non-whitespace chars
	if len(contents) > 2 || contents[1] != "" {
		fmt.Println("Warning, helper file ~FOLDERNAME should only contain 1 line")
	}
	return contents[0]
}

//CollectFSTree construct file trees according to
func CollectFSTree(path string, tracker *util.BookmarkTracker) *base.BookmarkNodeBase {
	// currently only handles URL

	toFileObj := func(file os.FileInfo, t times.Timespec, path string) *base.BookmarkNodeBase {
		var nodeType, URL, Name string
		if file.IsDir() {
			nodeType = "folder" // TODO could use enums here instead

			Name = getFolderName(path)

		} else if !file.IsDir() {
			nodeType = "url" // TODO could use enums here instead

			content, err := ioutil.ReadFile(path)
			check(err)

			contents := strings.SplitN(string(content), "\n", -1)
			//only first line contains non-whitespace chars
			if len(contents) < 3 {
				panic("Error, bookmark file did not contain lines for URL and name")
			}
			if len(contents) > 3 || contents[2] != "" {
				fmt.Println("Warning, only first two lines containing URL and name in file '" + path + "' were parsed.")
			}
			URL = contents[0]
			Name = contents[1]
			if Name == "~UNNAMED" {
				Name = ""
			}
		}

		node := base.BookmarkNodeBase{
			UUID:         uuid.New(),
			DateModified: int64(t.AccessTime().Unix()),
			DateCreated:  int64(t.ModTime().Unix()),
			Type:         nodeType,
			Name:         Name,
			URL:          URL,
			Path:         path,
			Children:     []*base.BookmarkNodeBase{},
		}

		key := util.TrackerKey{
			Name:    node.Name,
			Path:    node.Path,
			URL:     node.URL,
			Created: string(node.DateCreated),
		}

		tracker.Out[key] = path
		return &node
	}
	rootOSFile, _ := os.Stat(path)
	t := times.Get(rootOSFile)

	rootFile := toFileObj(rootOSFile, t, path) //start with root file
	stack := []*base.BookmarkNodeBase{rootFile}

	for len(stack) > 0 { //until stack is empty,
		file := stack[len(stack)-1] //pop entry from stack
		stack = stack[:len(stack)-1]
		dirFiles, _ := ioutil.ReadDir(file.Path) //get the children of entry
		for _, c := range dirFiles {             //for each child
			if c.Name() == "~FOLDERNAME" {
				continue
			}
			childPath := filepath.Join(file.Path, c.Name())
			t, err := times.Stat(childPath)
			check(err)
			child := toFileObj(c, t, childPath) //turn it into a base.BookmarkNodeBase object
			child.Parent = file
			file.Children = append(file.Children, child) //append it to the children of the current file popped
			stack = append(stack, child)                 //append the child to the stack, so the same process can be run again
		}
	}

	return rootFile
	// output, _ := json.MarshalIndent(rootFile, "", "     ")
	// fmt.Println(string(output))
}

//CollectFSTrees construct trees for root map
func CollectFSTrees(tmpRootPath string, tracker *util.BookmarkTracker) []*base.BookmarkNodeBase {
	return CollectFSTree(tmpRootPath, tracker).Children
}
