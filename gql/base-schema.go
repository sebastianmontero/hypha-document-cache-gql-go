package gql

import (
	"github.com/vektah/gqlparser/ast"
)

var DocumentFieldArgs = map[string]*SimplifiedField{
	"hash": {
		IsID:    true,
		Name:    "hash",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	"type": {
		Name:    "type",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	"creator": {
		Name:    "creator",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	"createdDate": {
		Name:    "createdDate",
		Type:    "DateTime",
		NonNull: true,
		Index:   "hour",
	},
}

const DocumentFields = `
		hash: String! @id @search(by: [exact])
		type: String! @search(by: [exact])
		creator: String! @search(by: [exact])
		createdDate: DateTime! @search(by: [hour])
`

const BaseSchema = `

	type Certificate {
		id: ID!
		certifier: String! @search(by: [exact])
		notes: String! @search(by: [term])
		certification_date: DateTime! @search(by: [hour])
	}

	interface Document {` +
	DocumentFields + `
	}

	type Cursor {
		id: String! @id @search(by: [exact])
		cursor: String!
	}

	type TypeVersion {
		type: String! @search(by: [exact])
		version: String @search(by: [exact])
	}
	
`

var CursorSimplifiedType = &SimplifiedType{
	Name: "Cursor",
	Fields: map[string]*SimplifiedField{
		"id": {
			Name:    "id",
			IsID:    true,
			Type:    "String",
			Index:   "exact",
			NonNull: true,
		},
		"cursor": {
			Name:    "cursor",
			Type:    "String",
			NonNull: true,
		},
	},
}

var BaseSchemaSource = &ast.Source{
	Input:   BaseSchema,
	BuiltIn: false,
}

func NewCursorInstance(id, cursor string) *SimplifiedInstance {
	return &SimplifiedInstance{
		SimplifiedType: CursorSimplifiedType,
		Values: map[string]interface{}{
			"id":     id,
			"cursor": cursor,
		},
	}
}
