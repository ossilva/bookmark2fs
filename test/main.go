package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

func translator(m map[rune]rune) func(rune) rune {
	return func(r rune) rune {
		val, ok := m[r]
		if ok {
			return val
		}
		return r
	}
}

var s2fn = translator(s2fnMap)
var s2fnMap = map[rune]rune{
	'<':  '﹤',
	'>':  '﹥',
	':':  'ː',
	'"':  '“',
	'/':  '⁄',
	'\\': '∖',
	'|':  '⼁',
	'?':  '﹖',
	'*':  '﹡',
	'.':  '⋅',
}

var fn2s = translator(fn2sMap)
var fn2sMap = map[rune]rune{
	'﹤': '<',
	'﹥': '>',
	'ː': ':',
	'“': '"',
	'⁄': '/',
	'∖': '\\',
	'⼁': '|',
	'﹖': '?',
	'﹡': '*',
	'⋅': '.',
}

//As above, but the minimum I needed for my files/filesystem/driver. Most
//notably, '<" and '>" seem to work fine, so there's no sense mangling them.
var s2fnMS = translator(s2fnMSMap)
var s2fnMSMap = map[rune]rune{
	':': 'ː',
	'?': '﹖',
	'|': '⼁',
	'/': '⁄',
	'.': '⋅',
}
var fn2sMS = translator(fn2sMSMap)
var fn2sMSMap = map[rune]rune{
	'ː': ':',
	'﹖': '?',
	'⼁': '|',
	'⁄': '/',
	'⋅': '.',
}

func StringToFilename(s string) string {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		return strings.Map(s2fnMS, s)
	default:
		return strings.Map(s2fn, s)
	}
}
func filenameToString(s string) string {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		return strings.Map(fn2sMS, s)
	default:
		return strings.Map(fn2s, s)
	}
}

type bookmarkNodeBase struct {
	Children     []*bookmarkNodeBase `json:"children"`
	DateCreated  int64               `json:"date_added"`
	DateModified int64               `json:"date_modified"`
	Name         string              `json:"name"`
	Type         string              `json:"type"`
	URL          string              `json:"url"`
	Path         string
	BookmarkBar  bool
	Baggage      *html.Node
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Tokenize HTML from stdin.
func main() {
	netscapeHTML, err := ioutil.ReadFile("./bookmarks_9_25_20.html")
	check(err)
	netscapeHTMLString := string(netscapeHTML)

	//remove redundant netscape '<p>' tags after '<dl>'
	r := regexp.MustCompile("(?i)<\\s*dl\\s*>\\s*<\\s*p\\s*>")
	htmlReader := strings.NewReader(r.ReplaceAllString(netscapeHTMLString, "<dl>"))

	doc, err := html.Parse(htmlReader)
	check(err)

	rootNodes := []*bookmarkNodeBase{}
	stack := []*bookmarkNodeBase{}

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

	for c := body.FirstChild; c != nil; c = c.NextSibling {
		if c.Data == "dt" {
			rootNode := &bookmarkNodeBase{}
			rootNode.Baggage = c
			stack = append(stack, rootNode)
			rootNodes = append(rootNodes, rootNode)
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
							node := &bookmarkNodeBase{}
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
					n.Path = path.Join(n.Path, StringToFilename(n.Name))
				}
			}
			n.Baggage = nil
		}
	}

	for _, node := range rootNodes {
		fmt.Println(node)
	}
}
