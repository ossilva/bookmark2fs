package jsonconv

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"time"

	"github.com/ossilva/bookmark2fs/pkg/conversion/base"
)

//BookmarkNodeJSON for parsing JSON
type BookmarkNodeJSON struct {
	Children           []*BookmarkNodeJSON `json:"children"`
	DateCreatedChrome  string              `json:"date_added"`
	DateModifiedChrome string              `json:"date_modified"`
	GUID               string              `json:"guid"` // TODO
	ID                 string              `json:"id"`   // TODO
	Name               string              `json:"name"`
	Type               string              `json:"type"`
	URL                string              `json:"url"`
	Path               string
}

// BookmarkJSON JSON folder roots
// JSON unmarshal cannot handle conversion to int
type BookmarkJSON struct {
	Checksum string                       `json:"checksum"` // md5?
	Roots    map[string]*BookmarkNodeJSON `json:"roots`
	Version  string                       `json:"version"`
	Unparsed map[string]interface{}       `json:",omitempty,remain"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// chrome uses webkit timestamps
func parseChromeTimeToUnix(chromeTimeStamp string) int64 {
	timeInt, err := strconv.ParseInt(chromeTimeStamp, 10, 64)
	if err != nil {
		return int64(0) //return unix epoch
	}
	// workaround for "time" package limitation of 250y durations
	divisor := float64(10)
	dividedTimeStamp := math.Round(float64(timeInt) / divisor)
	addString := strconv.FormatInt(int64(dividedTimeStamp), 10) + "us"
	toAdd, err := time.ParseDuration(addString)
	check(err)
	t := time.Date(1601, time.January, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < int(divisor); i++ {
		t = t.Add(toAdd)
	}
	return t.Unix()
}

// //defining UnmarshalJSON can also be used to convert between structs
// func (bmJSON *bookmarkJSON) UnmarshalJSON(data []byte) error {

//DecodeJSON parses JSON into abstract tree structs
func DecodeJSON(reader io.Reader) map[string]*base.BookmarkNodeBase {

	inJSONBytes, _ := ioutil.ReadAll(reader)

	var object BookmarkJSON

	json.Unmarshal(inJSONBytes, &object)

	var rootNodeMap = map[string]*base.BookmarkNodeBase{}
	for k, v := range object.Roots {
		rootNodeMap[k] = v.JSONNodeToBase()
	}
	return rootNodeMap
}

//JSONNodeToBase converts JSON to Base types. TODO incorporate in Marshal
func (j *BookmarkNodeJSON) JSONNodeToBase() *base.BookmarkNodeBase {
	var b base.BookmarkNodeBase

	b.DateCreated = parseChromeTimeToUnix(j.DateCreatedChrome)
	b.DateModified = parseChromeTimeToUnix(j.DateModifiedChrome)
	b.Name = j.Name
	b.Type = j.Type
	b.URL = j.URL

	var ms []*base.BookmarkNodeBase
	for _, c := range j.Children {
		ms = append(ms, c.JSONNodeToBase())
	}
	b.Children = ms
	return &b
}
