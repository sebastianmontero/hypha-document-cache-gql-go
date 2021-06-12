package domain

import (
	"fmt"
)

//Cursors helper to enable cursor decoding
type Cursors struct {
	Cursors []*Cursor `json:"cursors,omitempty"`
}

//Cursor domain object
type Cursor struct {
	UID    string   `json:"uid,omitempty"`
	Cursor string   `json:"cursor,omitempty"`
	DType  []string `json:"dgraph.type,omitempty"`
}

func (m *Cursor) String() string {
	return fmt.Sprintf("Cursor{UID: %v, Cursor: %v, DType: %v}", m.UID, m.Cursor, m.DType)
}
