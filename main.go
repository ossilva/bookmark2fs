package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	// "github.com/google/uuid"
	// "crypto/md5"
)

type bookmarkNode struct {
	Children  []bookmarkNode `json:"children"`
	DateAdded int            `json:"date_added"`
	GUID      string         `json:"guid"` // uuid
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	URL       string         `json:"url"`
}

type bmJSON struct {
	Checksum string                 `json:"checksum"` // md5?
	Roots    map[string]interface{} `json:"-"`
	Version  int                    `json:"version"`
}

func _main() {
	// Open our jsonFile
	jsonFile, err := os.Open("./bookmark_test.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var object bmJSON

	json.Unmarshal(byteValue, &object.Roots)
	if c, ok := object.Roots["checksum"].(string); ok {
		object.Checksum = string(c)
	}
	if v, ok := object.Roots["version"].(int); ok {
		object.Version = int(v)
	}
	delete(object.Roots, "checksum")
	delete(object.Roots, "version")

	roots := make(map[string]interface{})

	if r, ok := object.Roots["roots"].(map[string]interface{}); ok {
		for k, v := range r {
			rootsJSON, err := json.Marshal(v)
			if err != nil {
				// do error check
				fmt.Println(err)
				return
			}
			// fmt.Println(string(rootsJSON))
			var node bookmarkNode
			json.Unmarshal(rootsJSON, &node)
			fmt.Println(node)
			roots[k] = node
		}
	}

	fmt.Println(roots)
	fmt.Println(roots["synced"])

	// fmt.Println(object)
	// for i := 0; i < len(object.roots); i++ {
	// 	fmt.Println(object.roots)
	// }

	// res2D := &response2{
	// 	Page:   1,
	// 	Fruits: []string{"apple", "peach", "pear"}}
	// res2B, _ := json.Marshal(res2D)
	// fmt.Println(string(res2B))
	// // change field name formatting

	// var dat map[string]interface{}

	// data := det
	// for k, v := range data {
	// 	switch v := v.(type) {
	// 	case string:
	// 		fmt.Println(k, v, "(string)")
	// 	case float64:
	// 		fmt.Println(k, v, "(float64)")
	// 	case []interface{}:
	// 		fmt.Println(k, "(array):")
	// 		for i, u := range v {
	// 			fmt.Println("    ", i, u)
	// 		}
	// 	case interface{}:
	// 		fmt.Println(k, v.(), "(inteface)")
	// 	default:
	// 		fmt.Println(k, v, "(unknown)")
	// 	}
	// }
}

func main() {
	tmpl, err := template.ParseFiles("./bookmarks_9_25_20.html")
	if err != nil {
		// do error check
		fmt.Println(err)
		return
	}
	fmt.Println(tmpl)
}
