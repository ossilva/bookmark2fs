package util

import (
	"runtime"
	"strings"

	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
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

//StringToFilename convert original bookmark name to os path
func StringToFilename(s string) string {
	osName := runtime.GOOS
	var newFilename string
	switch osName {
	case "windows":
		newFilename = strings.Map(s2fnMS, s)
	default:
		newFilename = strings.Map(s2fn, s)
	}
	if len(newFilename) > 77 {
		newFilename = newFilename[:78] + "..."
	}
	return newFilename
}

//FilenameToString convert os path to original bookmark name
func FilenameToString(s string) string {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		return strings.Map(fn2sMS, s)
	default:
		return strings.Map(fn2s, s)
	}
}

type TrackerKey struct {
	Path, Name, URL, Created string
}

//BookmarkTracker is a simple wrapper for tracking input and output bookmarks
type BookmarkTracker struct {
	In  map[TrackerKey]string
	Out map[TrackerKey]string
}

//New constructor
func NewTracker() *BookmarkTracker {
	t := BookmarkTracker{
		In:  make(map[TrackerKey]string),
		Out: make(map[TrackerKey]string),
	}
	return &t
}

func (tracker *BookmarkTracker) Insert(node *base.BookmarkNodeBase) {
	key := TrackerKey{
		Name:    node.Name,
		Path:    node.Path,
		URL:     node.URL,
		Created: string(node.DateCreated),
	}
	tracker.In[key] = node.Path
}
