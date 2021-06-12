package gql_test

import (
	"fmt"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"gotest.tools/assert"
)

func TestAdd(t *testing.T) {

	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType := &gql.SimplifiedType{
		Name: "Assignment",
		Fields: map[string]*gql.SimplifiedField{
			"assignee": {
				Name:    "assignee",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
			"votes": {
				Name: "votes",
				Type: "Int64",
			},
		},
		ExtendsDocument: true,
	}
	hash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	expectedInstance := &gql.SimplifiedInstance{
		SimplifiedType: assignmentType,
		Values: map[string]interface{}{
			"hash":        hash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	}
	changed, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, changed, true)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Add(expectedInstance)
	assert.NilError(t, err)

	actualInstance, err := client.Get(hash, assignmentType)
	assert.NilError(t, err)

	fmt.Println("Actual Instance: ", actualInstance)

}
