// https://gist.github.com/lummie/7f5c237a17853c031a57277371528e87
package main

import (
	"bytes"
	"encoding/json"
)

type NodeType int

const (
	Folder NodeType = iota
	Url
)

var toString = map[NodeType]string{
	Folder: "folder",
	Url:    "url",
}

var toID = map[string]NodeType{
	"folder": Folder,
	"url":    Url,
}

func (n NodeType) String() string {
	return [...]string{"folder", "url"}[n]
}

func (n NodeType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(toString[n])
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

// UnmarshalJSON unmashals a quoted json string to the enum value
func (n *NodeType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	// Note that if the string cannot be found then it will be set to the zero value, 'Created' in this case.
	*n = toID[j]
	return nil
}
