package base

import "golang.org/x/net/html"

//BookmarkNodeBase for abstracting bookmark trees
type BookmarkNodeBase struct {
	Children     []*BookmarkNodeBase `json:"children"`
	DateCreated  int64               `json:"date_added"`
	DateModified int64               `json:"date_modified"`
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	URL          string              `json:"url"`
	Path         string
	BookmarkBar  bool
	Baggage      *html.Node
}
