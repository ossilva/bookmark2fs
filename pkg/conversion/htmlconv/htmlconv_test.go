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
	"1_empty_folder.html",
	"2_epiphany.html",
	"3_firefox2.html",
	"4_ie_sans_charset.html",
	"5_redditsaver.html",
	"6_whitespace_firefox.html",
	"7_duplicates.html",
	// firefox_bookmark_keywordHTML
}

func genNumMap(nums []int) map[string]int {
	numMap := make(map[string]int)
	for i, name := range testFilenames {
		numMap[name] = nums[i]
	}
	return numMap
}

func genFolderNumMap() map[string]int {
	folderNumMap := make(map[string]int)
	folderNums := []int{4, 0, 4, 0, 1, 1, 3}
	for i, name := range testFilenames {
		folderNumMap[name] = folderNums[i]
	}
	return folderNumMap
}

var folderNumMap map[string]int = genNumMap([]int{4, 0, 4, 0, 1, 1, 3}[:])
var nodeNumMap map[string]int = genNumMap([]int{6, 2, 6, 3, 3, 4, 5}[:])

func countFolders(node *base.BookmarkNodeBase) int {
	nodes := base.GetNodes(node)
	folderNum := 0
	for _, v := range nodes {
		if v.Type == "folder" {
			folderNum++
		}
	}
	return folderNum
}

func countFiles(node *base.BookmarkNodeBase) int {
	nodes := base.GetNodes(node)
	folderNum := 0
	for _, v := range nodes {
		if v.Type == "folder" {
			folderNum++
		}
	}
	return folderNum
}

func hasBookmarkBar(node *base.BookmarkNodeBase) bool {
	nodes := base.GetNodes(node)
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
func TestNodeNumber(t *testing.T) {
	for fname, folderNum := range nodeNumMap {
		reader, err := os.Open(fmt.Sprintf("./htmlconv_testdata/%s", fname))
		check(err)
		bookmarkRoots := ParseNetscapeHTML(reader)
		total := 0
		for _, root := range bookmarkRoots {
			total += len(base.GetNodes(root))
		}
		if total != folderNum {
			t.Errorf("Number of nodes was incorrect, got: %d, want: %d.", total, folderNum)
		}
	}
}

func TestFolderNumber(t *testing.T) {
	for fname, folderNum := range folderNumMap {
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
