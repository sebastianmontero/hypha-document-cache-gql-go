package gql

import (
	"github.com/vektah/gqlparser/ast"
)

var DocumentFieldArgs = []*SimplifiedField{
	{
		Name:    "hash",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	{
		Name:    "type",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	{
		Name:    "creator",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	{
		Name:    "createdDate",
		Type:    "DateTime",
		NonNull: true,
		Index:   "hour",
	},
}

const DocumentFields = `
		hash: String! @search(by: [exact])
		type: String! @search(by: [exact])
		creator: String! @search(by: [exact])
		createdDate: DateTime! @search(by: [hour])
`

const BaseSchema = `
	interface Document {` +
	DocumentFields + `
	}

	type TypeVersion {
		type: String! @search(by: [exact])
		version: String @search(by: [exact])
	}
	
`

var BaseSchemaSource = &ast.Source{
	Input:   BaseSchema,
	BuiltIn: false,
}
