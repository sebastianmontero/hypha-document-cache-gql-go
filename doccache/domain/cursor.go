package domain

import (
	"fmt"
)

//Cursors helper to enable cursor decoding
type Cursors struct {
	Cursors []*Cursor `json:"cursors,omitempty"`
}

//Represents the cursor as stored on the db, the cursor points to where we are on the delta stream
type Cursor struct {
	UID    string   `json:"uid,omitempty"`
	Cursor string   `json:"cursor,omitempty"`
	DType  []string `json:"dgraph.type,omitempty"`
}

func (m *Cursor) String() string {
	return fmt.Sprintf("Cursor{UID: %v, Cursor: %v, DType: %v}", m.UID, m.Cursor, m.DType)
}
