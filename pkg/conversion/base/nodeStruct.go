package base

import (
	"math/rand"

	"github.com/Masterminds/squirrel"
	"github.com/Masterminds/structable"
	"github.com/google/uuid"
	"golang.org/x/net/html"
)

//SupportedFields contains fields which are preserved in (de)serialization
var SupportedFields []string = []string{
	"DateCreated",
	"Name",
	"Type",
	"URL",
	"Path",
}

var bookmarkTableName = "bookmarks"

func GetNodes(node *BookmarkNodeBase) []*BookmarkNodeBase {
	nodes := []*BookmarkNodeBase{}
	stack := []*BookmarkNodeBase{node}
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

//BookmarkNodeBase for abstracting bookmark trees
type BookmarkNodeBase struct {
	UUID         uuid.UUID
	Children     []*BookmarkNodeBase `json:"children"`
	Parent       *BookmarkNodeBase   `json:"parent"`
	DateCreated  int64               `json:"date_added"`
	DateModified int64               `json:"date_modified"`
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	URL          string              `json:"url`
	Path         string
	BookmarkBar  bool
	Baggage      *html.Node
}

type RecordableNode struct {
	structable.Recorder
	builder      squirrel.StatementBuilderType
	ID           int64  `stbl:"id,PRIMARY_KEY"`
	UUID         string `stbl:"uuid"`
	ParentUUID   string `stbl:"parent_uuid"`
	DateCreated  int64  `stbl:"date_created"`
	DateModified int64  `stbl:"date_modified"`
	Name         string `stbl:"name"`
	Type         string `stbl:"type"`
	URL          string `stbl:"url"`
}

// ToRecordable converts base node to a recordable struct
func (node *BookmarkNodeBase) ToRecordable() *RecordableNode {
	var parentUUID string
	if node.Parent != nil {
		parentUUID = node.Parent.UUID.String()
	} else {
		parentUUID = ""
	}

	rec := new(RecordableNode)
	rec.ID = rand.Int63()
	rec.UUID = node.UUID.String()
	rec.ParentUUID = parentUUID
	rec.DateCreated = node.DateCreated
	rec.DateModified = node.DateModified
	rec.Name = node.Name
	rec.Type = node.Type
	rec.URL = node.URL

	return rec
}

func GetNodesBFS(bookmarks *BookmarkNodeBase) []*BookmarkNodeBase {
	nodes := []*BookmarkNodeBase{}
	queue := []*BookmarkNodeBase{bookmarks}

	for len(queue) > 0 {
		node := queue[0]
		if len(queue) > 1 {
			queue = queue[1:]
		} else {
			queue = queue[:0]
		}
		nodes = append(nodes, node)

		for _, c := range node.Children {
			queue = append(queue, c)
		}
	}
	return nodes
}
