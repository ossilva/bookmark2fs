package main

import (
	"runtime"
	"strings"
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

func stringToFilename(s string) string {
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
