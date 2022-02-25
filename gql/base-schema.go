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
		Index:   "regexp",
	},
	"createdDate": {
		Name:    "createdDate",
		Type:    "DateTime",
		NonNull: true,
		Index:   "hour",
	},
	"updatedDate": {
		Name:    "updatedDate",
		Type:    "DateTime",
		NonNull: true,
		Index:   "hour",
	},
	"contract": {
		Name:    "contract",
		Type:    "String",
		NonNull: true,
		Index:   "exact",
	},
}

const DocumentFields = `
		docId: String! @id @search(by: [exact])
		type: String! @search(by: [exact])
		creator: String! @search(by: [regexp])
		createdDate: DateTime! @search(by: [hour])
		updatedDate: DateTime! @search(by: [hour])
		contract: String! @search(by: [exact])
`

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

var DocumentSimplifiedInterface = &SimplifiedInterface{
	SimplifiedBaseType: &SimplifiedBaseType{
		Name:             "Document",
		Fields:           DocumentFieldArgs,
		WithSubscription: true,
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
