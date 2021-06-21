package gql_test

import (
	"fmt"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"gotest.tools/assert"
)

func TestAdd(t *testing.T) {
	beforeEach()
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
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
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
	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(personInstance.AddMutation(false))
	assert.NilError(t, err)

	actualPersonInstance, err := client.GetOne(personHash, personType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

}

func TestUpdate(t *testing.T) {
	beforeEach()
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
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	//***Update data only
	assignmentInstance = &gql.SimplifiedInstance{
		SimplifiedType: assignmentType,
		Values: map[string]interface{}{
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice1",
			"votes":       40,
		},
	}
	baseAssignmentType, err := schema.GetSimplifiedType("Assignment")
	assert.NilError(t, err)
	mutation, err := assignmentInstance.UpdateMutation(actualAssignmentInstance)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	//***Update schema and data
	assignmentType = &gql.SimplifiedType{
		Name: "Assignment",
		Fields: map[string]*gql.SimplifiedField{
			"assignee": {
				Name:    "assignee",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
			"periods": {
				Name: "periods",
				Type: "Int64",
			},
		},
		ExtendsDocument: true,
	}
	assignmentInstance = &gql.SimplifiedInstance{
		SimplifiedType: assignmentType,
		Values: map[string]interface{}{
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"periods":     11,
		},
	}
	updateOp, err = schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)

	baseAssignmentType, err = schema.GetSimplifiedType("Assignment")
	assert.NilError(t, err)
	mutation, err = assignmentInstance.UpdateMutation(actualAssignmentInstance)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne(assignmentHash, baseAssignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

}

func TestUpdateSetAddingDeletingEdge(t *testing.T) {
	beforeEach()
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
	personType := &gql.SimplifiedType{
		Name: "Person",
		Fields: map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
		},
		ExtendsDocument: true,
	}
	assignment1Hash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment1Instance := &gql.SimplifiedInstance{
		SimplifiedType: assignmentType,
		Values: map[string]interface{}{
			"hash":        assignment1Hash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	}

	personHash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	allPersonValues := map[string]interface{}{
		"hash":        personHash,
		"createdDate": "2020-11-12T18:27:47.000Z",
		"creator":     "dao.hypha",
		"type":        "person",
		"name":        "alice",
	}
	personInstance := &gql.SimplifiedInstance{
		SimplifiedType: personType,
		Values:         allPersonValues,
	}

	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)

	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignment1Instance.AddMutation(false))
	assert.NilError(t, err)
	err = client.Mutate(personInstance.AddMutation(false))
	assert.NilError(t, err)

	personType.Fields["assignments"] = gql.NewEdgeField("assignments", "Assignment")
	assignmentRef := map[string]interface{}{
		"hash": assignment1Hash,
	}
	setPersonValues := map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef},
	}
	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	_, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	mutation, err := personType.UpdateMutation(personHash, setPersonValues, nil)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err := client.GetOne(personHash, personType, nil)
	assert.NilError(t, err)

	allPersonValues["assignments"] = []map[string]interface{}{assignmentRef}
	personInstance = &gql.SimplifiedInstance{
		SimplifiedType: personType,
		Values:         allPersonValues,
	}

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

	assignment2Hash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment2Instance := &gql.SimplifiedInstance{
		SimplifiedType: assignmentType,
		Values: map[string]interface{}{
			"hash":        assignment2Hash,
			"createdDate": "2020-11-15T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "bob",
			"votes":       30,
		},
	}
	err = client.Mutate(assignment2Instance.AddMutation(false))
	assert.NilError(t, err)

	//***Add a new assignment to edge
	assignmentRef = map[string]interface{}{
		"hash": assignment2Hash,
	}
	setPersonValues = map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef},
	}
	mutation, err = personType.UpdateMutation(personHash, setPersonValues, nil)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err = client.GetOne(personHash, personType, nil)
	assert.NilError(t, err)
	allPersonValues["assignments"] = append(allPersonValues["assignments"].([]map[string]interface{}), assignmentRef)
	personInstance = &gql.SimplifiedInstance{
		SimplifiedType: personType,
		Values:         allPersonValues,
	}

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

	//***Remove an assignment from edge

	assignmentRef = map[string]interface{}{
		"hash": assignment1Hash,
	}
	removePersonValues := map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef},
	}

	mutation, err = personType.UpdateMutation(personHash, nil, removePersonValues)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err = client.GetOne(personHash, personType, nil)
	assert.NilError(t, err)

	allPersonValues["assignments"] = allPersonValues["assignments"].([]map[string]interface{})[1:]

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

}

func assertInstance(t *testing.T, actual, expected *gql.SimplifiedInstance) {
	assert.Assert(t, actual != nil, "Actual shouldn't be nil for expected: %v", expected)
	actualValues := filterNullValues(actual.Values)
	expectedValues := filterNullValues(expected.Values)
	assert.Equal(t, len(actualValues), len(expectedValues))
	simplifiedType := expected.SimplifiedType
	for name, value := range expectedValues {
		if fieldType, ok := simplifiedType.Fields[name]; ok && fieldType.IsObject() {
			expectedRefs := value.([]map[string]interface{})
			actualRefs := actualValues[name].([]interface{})
			// fmt.Println("Expected Refs: ", expectedRefs)
			// fmt.Println("Actual Refs: ", actualRefs)
			assert.Equal(t, len(expectedRefs), len(actualRefs))
			for _, expectedRef := range expectedRefs {
				assertContainsRef(t, actualRefs, expectedRef)
			}
		} else {
			assert.Equal(t, expectedValues[name], value)
		}
	}
}

func assertContainsRef(t *testing.T, actualRefs []interface{}, expectedRef map[string]interface{}) {
	expectedHash := expectedRef["hash"].(string)
	for _, actualRef := range actualRefs {
		if actualRef.(map[string]interface{})["hash"].(string) == expectedHash {
			return
		}
	}
	assert.Assert(t, false, fmt.Sprintf("Expected Hash: %v not founf", expectedHash))
}

func filterNullValues(values map[string]interface{}) map[string]interface{} {
	filtered := make(map[string]interface{})
	for key, value := range values {
		if value != nil {
			filtered[key] = value
		}
	}
	return filtered
}

func TestDelete(t *testing.T) {
	beforeEach()
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
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	mutation, err := assignmentInstance.DeleteMutation()
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)
	assert.Assert(t, actualAssignmentInstance == nil)

}

func TestMultipleMutations(t *testing.T) {
	beforeEach()
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
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)

	cursorId := "c1"
	cursorInstance := gql.NewCursorInstance(cursorId, "cursor1")

	err = client.Mutate(
		assignmentInstance.AddMutation(false),
		cursorInstance.AddMutation(true),
	)
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne(assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	actualCursorInstance, err := client.GetOne(cursorId, gql.CursorSimplifiedType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualCursorInstance, cursorInstance)

}
