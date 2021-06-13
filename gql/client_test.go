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
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := &gql.SimplifiedInstance{
		SimplifiedType: assignmentType,
		Values: map[string]interface{}{
			"hash":        assignmentHash,
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
	err = client.Add(assignmentInstance)
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.Get(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	personType := &gql.SimplifiedType{
		Name: "Person",
		Fields: map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
			"assignments": gql.NewEdgeField("assignments", "Assignment"),
		},
		ExtendsDocument: true,
	}
	personHash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	personInstance := &gql.SimplifiedInstance{
		SimplifiedType: personType,
		Values: map[string]interface{}{
			"hash":        personHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "person",
			"name":        "alice",
			"assignments": []map[string]interface{}{
				{
					"hash": assignmentHash,
				},
			},
		},
	}
	changed, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, changed, true)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Add(personInstance)
	assert.NilError(t, err)

	actualPersonInstance, err := client.Get(personHash, personType, nil)
	assert.NilError(t, err)

	fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

}

func assertInstance(t *testing.T, actual, expected *gql.SimplifiedInstance) {
	assert.Equal(t, len(actual.Values), len(expected.Values))
	simplifiedType := expected.SimplifiedType
	for name, value := range expected.Values {
		if fieldType, ok := simplifiedType.Fields[name]; ok && fieldType.IsObject() {
			expectedRefs := expected.Values[name].([]map[string]interface{})
			actualRefs := value.([]map[string]interface{})
			assert.Equal(t, len(expectedRefs), len(actualRefs))
			for _, expectedRef := range expectedRefs {
				assertContainsRef(t, actualRefs, expectedRef)
			}
		} else {
			assert.Equal(t, expected.Values[name], value)
		}
	}
}

func assertContainsRef(t *testing.T, actualRefs []map[string]interface{}, expectedRef map[string]interface{}) {
	expectedHash := expectedRef["hash"].(string)
	for _, actualRef := range actualRefs {
		if actualRef["hash"].(string) == expectedHash {
			return
		}
	}
	assert.Assert(t, false, fmt.Sprintf("Expected Hash: %v not founf", expectedHash))
}
