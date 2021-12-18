package gql

import (
	"github.com/vektah/gqlparser/ast"
)

var DocumentFieldArgs = map[string]*SimplifiedField{
	"docId": {
		IsID:    true,
		Name:    "docId",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
	"docId_i": {
		Name:    "docId_i",
		Type:    GQLType_Int64,
		NonNull: true,
		Index:   "int64",
	},
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
		docId: String! @id @search(by: [exact])
		docId_i: Int64! @search(by: [int64])
		hash: String! @id @search(by: [exact])
		type: String! @search(by: [exact])
		creator: String! @search(by: [exact])
		createdDate: DateTime! @search(by: [hour])
`

const BaseSchema = `

	type DocumentCertificate {
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

	type DoccacheConfig {
		id: String! @id @search(by: [exact])
		contract: String!
		eosEndpoint: String!
		documentsTable: String!
		edgesTable: String!
	}

	type TypeVersion {
		type: String! @search(by: [exact])
		version: String @search(by: [exact])
	}
	
`

var DocumentSimplifiedInterface = &SimplifiedInterface{
	SimplifiedBaseType: &SimplifiedBaseType{
		Name:   "Document",
		Fields: DocumentFieldArgs,
	},
}

var CursorSimplifiedType = &SimplifiedType{
	SimplifiedBaseType: &SimplifiedBaseType{
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
	},
}

var DoccacheConfigSimplifiedType = &SimplifiedType{
	SimplifiedBaseType: &SimplifiedBaseType{
		Name: "DoccacheConfig",
		Fields: map[string]*SimplifiedField{
			"id": {
				Name:    "id",
				IsID:    true,
				Type:    "String",
				Index:   "exact",
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
