package htmlconv

import (
	"fmt"
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

// func IsBookmarkNode(node) {
// 	s := regexp.MustCompile("(?i)\(dt\|\)")
// 	if s.FindString(htmlString) == "" {
// }

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

//ParseNetscapeHTML parses HTML to abstract trees
func ParseNetscapeHTML(reader io.Reader) []*base.BookmarkNodeBase {
	doc, err := html.Parse(reader)
	check(err)

	stack := []*base.BookmarkNodeBase{}

	var body *html.Node
	for d := doc.FirstChild; d != nil; d = d.NextSibling {
		if d.Type == html.ElementNode && (d.Data == "html" || d.Data == "body") {
			d = d.FirstChild
		}
		if d.Type == html.ElementNode && d.Data == "dl" {
			body = d
			break
		}
	}

	rootNodes := []*base.BookmarkNodeBase{}
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "dt" {
			rootNode := &base.BookmarkNodeBase{}
			rootNodes = append(rootNodes, rootNode)

			rootNode.UUID = uuid.New()
			rootNode.Baggage = c

			stack = append(stack, rootNode)
		}
	}

	for len(stack) > 0 { //until stack is empty,
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		bag := n.Baggage

		if bag.Type == html.ElementNode &&
			bag.Data == "dt" &&
			bag.FirstChild != nil {
			for dirCn := bag.FirstChild; dirCn != nil; dirCn = dirCn.NextSibling {
				if dirCn.Data == "dl" { //type: folder
					// dirCn = dirCn.FirstChild // children follow <p> tag
					for c := dirCn.FirstChild; c != nil; c = c.NextSibling {
						if c.Data == "p" {
							continue
						}
						if c.Data == "dt" {
							node := &base.BookmarkNodeBase{}
							node.UUID = uuid.New()
							node.Baggage = c
							node.Parent = n
							node.Path = n.Path
							n.Children = append(n.Children, node)
							stack = append(stack, node)
						}
					}
				} else { // if dirCn.Data == "h3" || dirCn.Data == "a" {
					switch dirCn.Data {
					case "h3": //type: folder
						n.Type = "folder"
						for _, a := range dirCn.Attr {
							if a.Key == "add_date" {
								n.DateCreated, err = strconv.ParseInt(a.Val, 10, 64)
								check(err)
								continue
							}
							if a.Key == "last_modified" {
								n.DateModified, err = strconv.ParseInt(a.Val, 10, 64)
								check(err)
								continue
							}
							if a.Key == "personal_toobar_folder" && a.Val == "true" {
								n.BookmarkBar = true
								continue
							}
						}
					case "a": //type: url
						n.Type = "url"
						for _, a := range dirCn.Attr {
							if a.Key == "add_date" {
								n.DateCreated, err = strconv.ParseInt(a.Val, 10, 64)
								check(err)
								continue
							}
							if a.Key == "href" {
								n.URL = a.Val
								continue
							}
						}
					default:
						continue
					}
					if dirCn.FirstChild != nil {
						n.Name = dirCn.FirstChild.Data
					} else {
						n.Name = "~UNNAMED"
					}
					n.Path = path.Join(n.Path, util.StringToFilename(n.Name))
					if n.DateCreated == 0 {
						fmt.Println("Error")
					}
				}
			}
			n.Baggage = nil
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
