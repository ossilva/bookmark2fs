package htmlconv

import (
	"fmt"
	"os"
	"testing"

	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
)

var testData []string = []string{
	"empty_folderHTML",
	"epiphanyHTML",
	"firefox2HTML",
	"ie_sans_charsetHTML",
	"redditsaverHTML",
	"whitespace_firefoxHTML",
	"duplicatesHTML",
}

var testFilenames []string = []string{
	"empty_folder.html",
	"epiphany.html",
	"firefox2.html",
	"ie_sans_charset.html",
	"redditsaver.html",
	"whitespace_firefox.html",
	"duplicates.html",
	// firefox_bookmark_keywordHTML
}

func genFolderNumMap() map[string]int {
	folderNumMap := make(map[string]int)
	folderNums := []int{4, 0, 4, 0, 1, 1, 3}
	for i, name := range testFilenames {
		folderNumMap[name] = folderNums[i]
	}
	return folderNumMap
}

var folderNumMap map[string]int = genFolderNumMap()

func getNodes(node *base.BookmarkNodeBase) []*base.BookmarkNodeBase {
	nodes := []*base.BookmarkNodeBase{}
	stack := []*base.BookmarkNodeBase{node}
	for len(stack) > 0 {
		v := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		nodes = append(nodes, v)
		for _, c := range v.Children {
			stack = append(stack, c)
		}
	}
	return nodes
}

func countFolders(node *base.BookmarkNodeBase) int {
	nodes := getNodes(node)
	folderNum := 0
	for _, v := range nodes {
		if v.Type == "folder" {
			folderNum++
		}
	}
	return folderNum
}

func countFiles(node *base.BookmarkNodeBase) int {
	nodes := getNodes(node)
	folderNum := 0
	for _, v := range nodes {
		if v.Type == "folder" {
			folderNum++
		}
	}
	return folderNum
}

func hasBookmarkBar(node *base.BookmarkNodeBase) bool {
	nodes := getNodes(node)
	for _, v := range nodes {
		if v.BookmarkBar {
			return true
		}
	}
	return false
}

func sum(arr []int) int {
	var s int
	for v := range arr {
		s += v
	}
	return s
}

func TestParseNotCausePanic(t *testing.T) {
	for _, v := range testFilenames {
		reader, err := os.Open(fmt.Sprintf("./htmlconv_testdata/%s", v))
		check(err)
		ParseNetscapeHTML(reader)
	}
}

func TestFolderNumber(t *testing.T) {
	// tst := map[string]int{"duplicates.html": 3}
	for fname, folderNum := range folderNumMap {
		// for fname, folderNum := range tst {
		reader, err := os.Open(fmt.Sprintf("./htmlconv_testdata/%s", fname))
		check(err)
		bookmarkRoots := ParseNetscapeHTML(reader)
		total := 0
		for _, root := range bookmarkRoots {
			total += countFolders(root)
		}
		if total != folderNum {
			t.Errorf("Number of folders was incorrect, got: %d, want: %d.", total, folderNum)
		}
	}
}

func TestParseNetscapeHTMLEmptyFolder(t *testing.T) {
	reader, err := os.Open("./htmlconv_testdata/empty_folder.html")
	check(err)
	bookmarkRoots := ParseNetscapeHTML(reader)
	total := 0
	for _, root := range bookmarkRoots {
		total += countFolders(root)
	}
	t.Errorf("Number of folders was incorrect, got: %d, want: %d.", total, 4)
}

func TestDuplicateHTMLFolderNum(t *testing.T) {
	reader, err := os.Open("./htmlconv_testdata/duplicates.html")
	check(err)
	bookmarkRoots := ParseNetscapeHTML(reader)
	total := 0
	for _, root := range bookmarkRoots {
		total += countFolders(root)
	}
	t.Errorf("Number of folders was incorrect, got: %d, want: %d.", total, 3)
}

func TestDuplicateHTMLFileNum(t *testing.T) {
	reader, err := os.Open("./htmlconv_testdata/duplicates.html")
	check(err)
	bookmarkRoots := ParseNetscapeHTML(reader)
	total := 0
	for _, root := range bookmarkRoots {
		total += countFiles(root)
	}
	t.Errorf("Number of folders was incorrect, got: %d, want: %d.", total, 3)
}
