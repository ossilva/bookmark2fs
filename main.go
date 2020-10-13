package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	// "github.com/google/uuid"
	// "crypto/md5"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// chrome uses webkit timestamps
func chromeToUnixTime(chromeTimeStamp string) int64 {
	timeInt, err := strconv.ParseInt(chromeTimeStamp, 10, 64)
	if err != nil {
		return int64(0)
	}
	divisor := float64(10)
	t := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	dividedTimeStamp := math.Round(float64(timeInt) / divisor)
	addString := strconv.FormatInt(int64(dividedTimeStamp), 10) + "us"
	toAdd, err := time.ParseDuration(addString)
	check(err)
	for i := 0; i < int(divisor); i++ {
		t = t.Add(toAdd)
	}
	unixTime := t.Unix()
	return unixTime
}

func JSONDecoder(targetObj *bookmarkJSON) *mapstructure.Decoder {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           targetObj,
	}
	decoder, err := mapstructure.NewDecoder(config)
	check(err)

	return decoder
}

type bookmarkNodeJSON struct {
	Children           []*bookmarkNodeJSON `json:"children"`
	DateCreatedChrome  string              `json:"date_added"`
	DateModifiedChrome string              `json:"date_modified"`
	GUID               string              `json:"guid"` // TODO
	ID                 string              `json:"id"`   // TODO
	Name               string              `json:"name"`
	Type               string              `json:"type"`
	URL                string              `json:"url"`
	Path               string
	// Unparsed           map[string]interface{} `json:",remain"` // debug
}

type bookmarkNodeBase struct {
	Children     []*bookmarkNodeBase `json:"children"`
	DateCreated  int64               `json:"date_added"`
	DateModified int64               `json:"date_modified"`
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	URL          string              `json:"url"`
	Path         string
	// Unparsed     map[string]interface{} `json:",remain"` // debug
}

type bookmarkRootBase struct {
	Checksum string                       `json:"checksum"` // md5?
	Roots    map[string]*bookmarkNodeBase `json:"roots`
	Version  string                       `json:"version"`
}

type bookmarkJSON struct {
	Checksum string                       `json:"checksum"` // md5?
	Roots    map[string]*bookmarkNodeJSON `json:"roots`
	Version  string                       `json:"version"`
	Unparsed map[string]interface{}       `json:",omitempty,remain"`
}

// convert JSON to Base types
func (j *bookmarkNodeJSON) jsonNodeToBase() *bookmarkNodeBase {
	var b bookmarkNodeBase

	b.DateCreated = chromeToUnixTime(j.DateCreatedChrome)
	b.DateModified = chromeToUnixTime(j.DateModifiedChrome)
	b.Name = j.Name
	b.Type = j.Type
	b.URL = j.URL

	var baseBms []*bookmarkNodeBase
	for _, c := range j.Children {
		baseBms = append(baseBms, c.jsonNodeToBase())
	}
	b.Children = baseBms
	return &b
}

func readUserInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("type \"done\" to save new bookmarks")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if strings.Compare("done", text) == 0 {
			return
		}

	}
}

func populateTmpDir(baseDirs map[string]*bookmarkNodeBase) (string, map[string]string) {
	programName := "bookmark2fs"
	tmpDirPath, err := ioutil.TempDir(os.TempDir(), programName)
	check(err)

	populateRootDir := func(rootNode *bookmarkNodeBase) string {
		stack := []*bookmarkNodeBase{rootNode}
		filename := stringToFilename(rootNode.Name)
		rootNode.Path = path.Join(tmpDirPath, filename)
		for len(stack) > 0 {
			fileNode := stack[len(stack)-1]
			nodePath := fileNode.Path
			stack = stack[:len(stack)-1]

			switch fileNode.Type {
			case "url":
				d := []byte(fileNode.URL + "\n")

				err := ioutil.WriteFile(fileNode.Path, d, 0644)
				check(err)
				ctime := time.Unix(fileNode.DateCreated, 0)
				atime := time.Unix(fileNode.DateModified, 0)
				err = os.Chtimes(fileNode.Path, atime, ctime)
				check(err)

			case "folder":
				err := os.Mkdir(nodePath, 0755)
				check(err)

				// defer os.RemoveAll(nodePath)

				for _, child := range fileNode.Children {
					childFName := stringToFilename(child.Name)
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

func buildFileTreeHTML(bookmarkTreeData *bookmarkRootBase, outPath string) {
	tmpl, err := template.ParseFiles("./netscape_bookmarks.tmpl")
	f, err := os.Create(outPath)
	check(err)
	defer f.Close()

	w := bufio.NewWriter(f)
	tmpl.ExecuteTemplate(w, "main", bookmarkTreeData)
	w.Flush()
}

func main() {
	outPath := "/tmp/test_bms_out.html"
	jsonFile, err := os.Open("./bookmark_test.json")

	check(err)
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var object bookmarkJSON

	// decoder := JSONDecoder(&object)

	// err = decoder.Decode(byteValue)
	// check(err)

	cloneRootMetadata := func(orig bookmarkJSON) *bookmarkRootBase {
		return &bookmarkRootBase{
			Checksum: orig.Checksum,
			Roots:    map[string]*bookmarkNodeBase{},
			Version:  orig.Version,
		}
	}

	json.Unmarshal(byteValue, &object)
	var bookmarkRoot = cloneRootMetadata(object)
	for k, v := range object.Roots {
		bookmarkRoot.Roots[k] = v.jsonNodeToBase()
	}

	// for k, v := range object.Unparsed {
	// for k, v := range object.Roots {
	// 	rootsJSON, err := json.Marshal(v)
	// 	if err != nil {
	// 		// do error check
	// 		fmt.Println(err)
	// 		return
	// 	}
	// 	var node *bookmarkNodeJSON
	// 	json.Unmarshal(rootsJSON, &node)
	// 	bookmarkRoot.Roots[k] = node.jsonNodeToBase()
	// }

	tmpDirPath, tmpRootPaths := populateTmpDir(bookmarkRoot.Roots)

	var exportRoot = cloneRootMetadata(object)
	for k := range object.Roots {
		rootPath := tmpRootPaths[k]
		exportRoot.Roots[k] = constructFSTree(rootPath)
	}
	buildFileTreeHTML(exportRoot, outPath)

	readUserInput()
	defer os.RemoveAll(tmpDirPath)
}

// func convertJSONRootToBase(rootDirsJSON *bookmarkNodeJSON) *bookmarkNodeBase {
// 	rootDirsBase := make(map[string]*bookmarkNodeBase)
// 	for k, v := range rootDirsJSON {
// 		rootDirsBase[k] = v.jsonNodeToBase()
// 	}
// 	return rootDirsBase
// }

// func convertJSONRootsToBase(rootDirsJSON map[string]*bookmarkNodeJSON) map[string]*bookmarkNodeBase {
// 	rootDirsBase := make(map[string]*bookmarkNodeBase)
// 	for k, v := range rootDirsJSON {
// 		rootDirsBase[k] = v.jsonNodeToBase()
// 	}
// 	return rootDirsBase
// }
