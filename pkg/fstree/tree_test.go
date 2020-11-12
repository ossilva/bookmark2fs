package fstree

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/util"
)

// https://stackoverflow.com/questions/18930910/access-struct-property-by-name
func getField(b *base.BookmarkNodeBase, field string) int {
	r := reflect.ValueOf(b)
	f := reflect.Indirect(r).FieldByName(field)
	return int(f.Int())
}

func nodesSimilar(n1 *base.BookmarkNodeBase, n2 *base.BookmarkNodeBase) bool {
	// for _, f := range base.SupportedFields {
	// 	if getField(n1, f) != getField(n2, f) {
	// 		return false
	// 	}
	// }
	if n1.Name == n2.Name &&
		n1.Type == n2.Type &&
		n1.URL == n2.URL &&
		n1.Path == n2.Path {
		return true
	}
	return false
}

func TestCollectTrees(t *testing.T) {

	testURLNode := &base.BookmarkNodeBase{
		DateCreated: 1505127380,
		Name:        "testBookmark",
		Type:        "url",
		URL:         "http://wikipedia.org",
		Path:        "testBase/testBookmark",
	}

	testFolderNode := &base.BookmarkNodeBase{
		DateCreated: 1605127380,
		Children:    []*base.BookmarkNodeBase{testURLNode},
		Name:        "testBookmarkFolder",
		Type:        "folder",
		Path:        "testBase",
	}

	testPath := "/tmp/bm2fsTest"
	defer os.RemoveAll(testPath)
	tracker := util.NewTracker()

	testTree := []*base.BookmarkNodeBase{testFolderNode}
	testNodes := base.GetNodes(testFolderNode)
	PopulateTmpDir(testTree, tracker, testPath)
	collectedTrees := CollectFSTrees(testPath, tracker)

	if len(testTree) != len(collectedTrees) {
		t.Errorf("Temp root dir collection failed")
	}

	collectedNodes := base.GetNodes(collectedTrees[0])

	var differingNodes []string = []string{}
	for i := range collectedTrees {
		n1, n2 := testNodes[i], collectedNodes[i]
		eq := nodesSimilar(n1, n2)
		if !eq {
			differingNodes = append(differingNodes, n1.Name)
			fmt.Println(n1.DateCreated, n2.DateCreated)
		}
	}
	if len(differingNodes) > 0 {
		t.Errorf(fmt.Sprintf("Conversion for the following nodes %v", differingNodes))
	}

}
