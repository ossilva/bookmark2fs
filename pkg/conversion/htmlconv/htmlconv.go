package htmlconv

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
	"github.com/ossilva/bookmark2fs/pkg/util"
	"golang.org/x/net/html"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

//BuildTreeHTML encodes abstract trees as html
func BuildTreeHTML(roots map[string]*base.BookmarkNodeBase, outPath string) {
	tmpl, err := template.ParseFiles("./netscape_bookmarks.tmpl")
	f, err := os.Create(outPath)
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	tmpl.ExecuteTemplate(w, "main", roots)
	w.Flush()
}

//ParseNetscapeHTML parses HTML to abstract trees
func ParseNetscapeHTML(reader io.Reader) map[string]*base.BookmarkNodeBase {
	netscapeHTML, err := ioutil.ReadFile("./bookmarks_9_25_20.html")
	check(err)
	netscapeHTMLString := string(netscapeHTML)

	//remove redundant netscape '<p>' tags after '<dl>'
	r := regexp.MustCompile("(?i)<\\s*dl\\s*>\\s*<\\s*p\\s*>")
	htmlReader := strings.NewReader(r.ReplaceAllString(netscapeHTMLString, "<dl>"))

	doc, err := html.Parse(htmlReader)
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

	rootNodeMap := map[string]*base.BookmarkNodeBase{}
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "dt" {
			rootNode := &base.BookmarkNodeBase{}
			rootNode.Baggage = c
			rootName := rootNode.Baggage.FirstChild.Data
			isBmBar := false
			for _, a := range c.Attr {
				isBmBar = isBmBar ||
					a.Key == "PERSONAL_TOOLBAR_FOLDER" &&
						a.Val == "true"
			}
			if isBmBar {
				rootName = "Bookmarks bar"
			}

			stack = append(stack, rootNode)
			rootNodeMap[rootName] = rootNode
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
					fmt.Println("dl")
					// dirCn := nodeC.FirstChild // children enclosed in <p></p>
					for c := dirCn.FirstChild; c != nil; c = c.NextSibling {
						if c.Data == "dt" {
							node := &base.BookmarkNodeBase{}
							node.Baggage = c
							node.Path = n.Path
							n.Children = append(n.Children, node)
							stack = append(stack, node)
						}
					}
				} else if dirCn.Data == "h3" || dirCn.Data == "a" {
					switch dirCn.Data {
					case "h3":
						n.Type = "file"
						for _, a := range dirCn.Attr {
							if a.Key == "add_date" {
								n.DateCreated, err = strconv.ParseInt(a.Val, 10, 64)
								check(err)
								break
							}
							if a.Key == "last_modified" {
								n.DateModified, err = strconv.ParseInt(a.Val, 10, 64)
								check(err)
								break
							}
							if a.Key == "personal_toobar_folder" && a.Val == "true" {
								n.BookmarkBar = true
								break
							}
						}
					case "a": //type: file
						n.Type = "folder"
						for _, a := range dirCn.Attr {
							if a.Key == "add_date" {
								n.DateCreated, err = strconv.ParseInt(a.Val, 10, 64)
								check(err)
								break
							}
							if a.Key == "href" {
								n.URL = a.Val
								break
							}
						}
					}
					n.Name = dirCn.FirstChild.Data
					n.Path = path.Join(n.Path, util.StringToFilename(n.Name))
				}
			}
			n.Baggage = nil
		}
	}
	return rootNodeMap
}
