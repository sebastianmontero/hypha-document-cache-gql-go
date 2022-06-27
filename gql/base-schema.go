package gql

import (
	"github.com/vektah/gqlparser/ast"
)

// Defines the structs that represent the objects that are part of the initial schema
// of every document cache instance

//The fields that are part of the document interface
var DocumentFieldArgs = map[string]*SimplifiedField{
	"docId": {
		IsID:    true,
		Name:    "docId",
		Type:    "String",
		NonNull: true,
		Indexes: NewIndexes("exact"),
	},
	"type": {
		Name:    "type",
		Type:    "String",
		NonNull: true,
		Indexes: NewIndexes("exact"),
	},
	"creator": {
		Name:    "creator",
		Type:    "String",
		NonNull: true,
		Indexes: NewIndexes("exact", "regexp"),
	},
	"createdDate": {
		Name:    "createdDate",
		Type:    "DateTime",
		NonNull: true,
		Indexes: NewIndexes("hour"),
	},
	"updatedDate": {
		Name:    "updatedDate",
		Type:    "DateTime",
		NonNull: true,
		Indexes: NewIndexes("hour"),
	},
	"contract": {
		Name:    "contract",
		Type:    "String",
		NonNull: true,
		Indexes: NewIndexes("exact"),
	},
}

const DocumentFields = `
		docId: String! @id @search(by: [exact])
		type: String! @search(by: [exact])
		creator: String! @search(by: [exact, regexp])
		createdDate: DateTime! @search(by: [hour])
		updatedDate: DateTime! @search(by: [hour])
		contract: String! @search(by: [exact])
`

// The base graphql schema
const BaseSchema = `

	interface Document @withSubscription {` +
	DocumentFields + `
	}

	type Cursor {
		id: String! @id @search(by: [exact])
		cursor: String!
	}

	type DoccacheConfig {
		id: String! @id @search(by: [exact])
		contract: String!
		eosEndpoint: String!
		documentsTable: String!
		edgesTable: String!
		elasticEndpoint: String!
		elasticApiKey: String!
	}

	type TypeVersion {
		type: String! @search(by: [exact])
		version: String @search(by: [exact])
	}
	
`

// The document interface type
var DocumentSimplifiedInterface = &SimplifiedInterface{
	SimplifiedBaseType: &SimplifiedBaseType{
		Name:             "Document",
		Fields:           DocumentFieldArgs,
		WithSubscription: true,
	},
}

// The cursor type
var CursorSimplifiedType = &SimplifiedType{
	SimplifiedBaseType: &SimplifiedBaseType{
		Name: "Cursor",
		Fields: map[string]*SimplifiedField{
			"id": {
				Name:    "id",
				IsID:    true,
				Type:    "String",
				Indexes: NewIndexes("exact"),
				NonNull: true,
			},
			"cursor": {
				Name:    "cursor",
				Type:    "String",
				NonNull: true,
			},
		},
	},
}

// The document cache config type
var DoccacheConfigSimplifiedType = &SimplifiedType{
	SimplifiedBaseType: &SimplifiedBaseType{
		Name: "DoccacheConfig",
		Fields: map[string]*SimplifiedField{
			"id": {
				Name:    "id",
				IsID:    true,
				Type:    "String",
				Indexes: NewIndexes("exact"),
				NonNull: true,
			},
			"contract": {
				Name:    "contract",
				Type:    "String",
				NonNull: true,
			},
			"eosEndpoint": {
				Name:    "eosEndpoint",
				Type:    "String",
				NonNull: true,
			},
			"documentsTable": {
				Name:    "documentsTable",
				Type:    "String",
				NonNull: true,
			},
			"edgesTable": {
				Name:    "edgesTable",
				Type:    "String",
				NonNull: true,
			},
			"elasticEndpoint": {
				Name:    "elasticEndpoint",
				Type:    "String",
				NonNull: true,
			},
			"elasticApiKey": {
				Name:    "elasticApiKey",
				Type:    "String",
				NonNull: true,
			},
		},
	},
}

var BaseSchemaSource = &ast.Source{
	Input:   BaseSchema,
	BuiltIn: false,
}

func NewCursorInstance(id, cursor string) *SimplifiedInstance {
	return NewSimplifiedInstance(
		CursorSimplifiedType,
		map[string]interface{}{
			"id":     id,
			"cursor": cursor,
		},
	)
}
