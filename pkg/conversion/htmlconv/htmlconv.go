package htmlconv

import (
	"errors"
	"io"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/google/uuid"
	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/util"
	"golang.org/x/net/html"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//BuildTreeHTML serializes abstract trees as template Netscape HTML
func BuildTreeHTML(roots []*base.BookmarkNodeBase, outPath string) {
	tmpl, err := template.ParseFiles("./netscape_bookmarks.tmpl")
	check(err)
	var w io.Writer
	if outPath == "stdout" {
		w = os.Stdout
	} else {
		var err error
		w, err = os.Create(outPath)
		check(err)
		defer w.(*os.File).Close()
	}
	writeErr := tmpl.ExecuteTemplate(w, "main", roots)
	check(writeErr)
}

func getFolderName(node *html.Node) string {
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "h3" {
			if cc := c.FirstChild; cc != nil {
				return cc.Data
			}
		}
	}
	return ""
}

// maybe should handle like <dt>
func standardizeHTML(htmlString string) *strings.Reader {
	s := regexp.MustCompile("(?i)<\\s*/p\\s*>")

	var htmlReader *strings.Reader
	//remove redundant netscape '<p>' tags after '<dl>'
	if s.FindString(htmlString) == "" {
		r := regexp.MustCompile("(?i)<\\s*dl\\s*>\\s*<\\s*p\\s*>")
		htmlReader = strings.NewReader(r.ReplaceAllString(htmlString, "<dl>"))
	} else {
		htmlReader = strings.NewReader(htmlString)
	}
	return htmlReader
}

func locateGroundNode(doc *html.Node) (*html.Node, error) {
	for d := doc.FirstChild; d != nil; d = d.NextSibling {
		if d.Type == html.ElementNode && (d.Data == "html" || d.Data == "body") {
			d = d.FirstChild
		}
		if d.Type == html.ElementNode && d.Data == "dl" {
			return d, nil
		}
	}
	return nil, errors.New("Could not find bottom layer of bookmark html.")
}

func collectRoots(groundNode *html.Node) []*base.BookmarkNodeBase {
	rootNodes := []*base.BookmarkNodeBase{}
	for c := groundNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "dt" {
			rootNode := &base.BookmarkNodeBase{}
			rootNodes = append(rootNodes, rootNode)

			rootNode.UUID = uuid.New()
			rootNode.Baggage = c
		}
	}
	return rootNodes
}

func processDir(
	stack []*base.BookmarkNodeBase,
	dirNode *html.Node,
	bmNode *base.BookmarkNodeBase,
) []*base.BookmarkNodeBase {
	// dirCn = dirCn.FirstChild // children follow <p> tag
	newStack := stack
	for c := dirNode.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "p" {
			continue
		}
		if c.Data == "dt" {
			dirNode := &base.BookmarkNodeBase{}
			dirNode.UUID = uuid.New()
			dirNode.Baggage = c
			dirNode.Parent = bmNode
			dirNode.Path = bmNode.Path
			bmNode.Children = append(bmNode.Children, dirNode)
			newStack = append(newStack, dirNode)
		}
	}
	return newStack
}

func annotateDirNode(bmNode *base.BookmarkNodeBase, dirNode *html.Node) *base.BookmarkNodeBase {
	bmNode.Type = "folder"
	var err error
	for _, a := range dirNode.Attr {
		if a.Key == "add_date" {
			bmNode.DateCreated, err = strconv.ParseInt(a.Val, 10, 64)
		} else if a.Key == "last_modified" {
			bmNode.DateModified, err = strconv.ParseInt(a.Val, 10, 64)
		} else if a.Key == "personal_toobar_folder" && a.Val == "true" {
			bmNode.BookmarkBar = true
		}
		check(err)
	}
	return bmNode
}

func annotateUrlNode(bmNode *base.BookmarkNodeBase, dirNode *html.Node) *base.BookmarkNodeBase {
	bmNode.Type = "url"
	var err error
	for _, a := range dirNode.Attr {
		if a.Key == "add_date" {
			bmNode.DateCreated, err = strconv.ParseInt(a.Val, 10, 64)
			check(err)
		} else if a.Key == "href" {
			bmNode.URL = a.Val
		}
	}
	return bmNode
}

//ParseNetscapeHTML parses HTML to abstract trees
func ParseNetscapeHTML(reader io.Reader) []*base.BookmarkNodeBase {
	doc, err := html.Parse(reader)
	check(err)

	groundNode, err := locateGroundNode(doc)
	check(err)

	rootNodes := collectRoots(groundNode)
	stack := append([]*base.BookmarkNodeBase{}, rootNodes...)

	for len(stack) > 0 { //until stack is empty,
		bmNode := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		bag := bmNode.Baggage

		if bag.Type == html.ElementNode &&
			bag.Data == "dt" &&
			bag.FirstChild != nil {
			for dirChild := bag.FirstChild; dirChild != nil; dirChild = dirChild.NextSibling {
				if dirChild.Data == "dl" { //type: folder
					stack = processDir(stack, dirChild, bmNode)
				} else { // if dirCn.Data == "h3" || dirCn.Data == "a" {
					switch dirChild.Data {
					case "h3": //type: folder
						bmNode = annotateDirNode(bmNode, dirChild)
					case "a": //type: url
						bmNode = annotateUrlNode(bmNode, dirChild)
					default:
						continue
					}
					if dirChild.FirstChild != nil {
						bmNode.Name = dirChild.FirstChild.Data
					} else {
						bmNode.Name = "~UNNAMED"
					}
					bmNode.Path = path.Join(bmNode.Path, util.StringToFilename(bmNode.Name))
				}
			}
			bmNode.Baggage = nil
		}
	}
	return rootNodes
}

func skipPTag(node *html.Node) *html.Node {
	if node.Data == "p" {
		return node.FirstChild
	}
	return node
}
