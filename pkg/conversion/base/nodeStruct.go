package base

import (
	"github.com/google/uuid"
	"github.com/ossilva/bookmark2fs/pkg/db"
	"golang.org/x/net/html"
)

//SupportedFields contains fields which are preserved in (de)serialization
var SupportedFields []string = []string{
	// "DateCreated",
	"Name",
	"Type",
	"URL",
	"Path",
}

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

func (node *BookmarkNodeBase) ToRecordable() *db.RecordableNode {
	rec := &db.RecordableNode{
		ID:           node.UUID.String(),
		ParentId:     node.Parent.UUID.String(),
		DateCreated:  node.DateCreated,
		DateModified: node.DateModified,
		Name:         node.Name,
		Type:         node.Type,
		URL:          node.URL,
	}
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
		return nodes
	}

	return nodes
}
