package gql_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"gotest.tools/assert"
)

func TestAdd(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType := gql.NewSimplifiedType(

		"Assignment",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentId := "1"
	assignmentIdI, _ := strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	personType := gql.NewSimplifiedType(

		"Person",
		map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
			"assignments": gql.NewEdgeField("assignments", "Assignment"),
		},
		gql.DocumentSimplifiedInterface,
	)
	personId := "2"
	personIdI, _ := strconv.ParseUint(personId, 10, 64)
	personHash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	personInstance := gql.NewSimplifiedInstance(
		personType,
		map[string]interface{}{
			"docId":       personId,
			"docId_i":     personIdI,
			"hash":        personHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "person",
			"name":        "alice",
			"assignments": []map[string]interface{}{
				{
					"docId": assignmentId,
				},
			},
		},
	)
	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(personInstance.AddMutation(false))
	assert.NilError(t, err)

	actualPersonInstance, err := client.GetOne("hash", personHash, personType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

}

func TestAddMultipleIDs(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				IsID:    true,
			},
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentId := "1"
	assignmentIdI, _ := strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"name":        "assign1",
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	// Verify that each id field is independent, not composite id
	assertInstance(t, actualAssignmentInstance, assignmentInstance)
	assignmentId = "2"
	assignmentIdI, _ = strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash = "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951f"
	assignmentInstance.SetValue("docId", assignmentId)
	assignmentInstance.SetValue("docId_i", assignmentIdI)
	assignmentInstance.SetValue("hash", assignmentHash)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.ErrorContains(t, err, "id assign1 already exists for field name inside type Assignment")

	assignmentInstance.SetValue("name", "assign2")
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

}

func TestUpdateFieldToBeID(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)

	assignmentType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
			"name": {
				Name:  "name",
				Type:  "String",
				Index: "term",
			},
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentId := "1"
	assignmentIdI, _ := strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"name":        "assign1",
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	// Verify that each id field is independent, not composite id
	assertInstance(t, actualAssignmentInstance, assignmentInstance)
	assignmentId = "2"
	assignmentIdI, _ = strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash = "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951f"
	assignmentInstance.SetValue("docId", assignmentId)
	assignmentInstance.SetValue("docId_i", assignmentIdI)
	assignmentInstance.SetValue("hash", assignmentHash)
	assignmentInstance.SetValue("name", "assign2")
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	assignmentType.Fields["name"] = &gql.SimplifiedField{

		Name:    "name",
		Type:    "String",
		IsID:    true,
		NonNull: true,
		Index:   "term",
	}

	updateOp, err = schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	actualAssignmentInstance, err = client.GetOne("name", "assign2", assignmentType, nil)
	assert.NilError(t, err)

	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	// currentSchema, err := admin.GetCurrentSchema()
	// assert.NilError(t, err)
	// aType := currentSchema.GetType("Assignment")
	// fmt.Println(gql.DefinitionToString(aType, 0))
	// sType, err := gql.NewSimplifiedTypeFromType(aType)
	// assert.NilError(t, err)
	// fmt.Println("sType: ", sType)

}

func TestUpdate(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType := gql.NewSimplifiedType(

		"Assignment",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentId := "1"
	assignmentIdI, _ := strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	//***Update data only
	assignmentInstance = gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice1",
			"votes":       40,
		},
	)
	baseAssignmentType, err := schema.GetSimplifiedType("Assignment")
	assert.NilError(t, err)
	mutation, err := assignmentInstance.UpdateMutation("hash", actualAssignmentInstance)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne("hash", assignmentHash, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	//***Update schema and data
	assignmentType = gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentInstance = gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"periods":     11,
		},
	)
	updateOp, err = schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)

	baseAssignmentType, err = schema.GetSimplifiedType("Assignment")
	assert.NilError(t, err)
	mutation, err = assignmentInstance.UpdateMutation("hash", actualAssignmentInstance)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne("hash", assignmentHash, baseAssignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

}

func TestUpdateSetAddingDeletingEdge(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)
	personType := gql.NewSimplifiedType(
		"Person",
		map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
		},
		gql.DocumentSimplifiedInterface,
	)
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Hash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment1Instance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignment1Id,
			"docId_i":     assignment1IdI,
			"hash":        assignment1Hash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
	personId := "2"
	personIdI, _ := strconv.ParseUint(personId, 10, 64)
	personHash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	allPersonValues := map[string]interface{}{
		"docId":       personId,
		"docId_i":     personIdI,
		"hash":        personHash,
		"createdDate": "2020-11-12T18:27:47.000Z",
		"creator":     "dao.hypha",
		"type":        "person",
		"name":        "alice",
	}
	personInstance := gql.NewSimplifiedInstance(
		personType,
		allPersonValues,
	)

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
		"docId": assignment1Id,
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
	mutation, err := personType.UpdateMutation("docId", personId, setPersonValues, nil)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err := client.GetOne("docId", personId, personType, nil)
	assert.NilError(t, err)

	allPersonValues["assignments"] = []map[string]interface{}{assignmentRef}
	personInstance = gql.NewSimplifiedInstance(
		personType,
		allPersonValues,
	)

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

	assignment2Id := "3"
	assignment2IdI, _ := strconv.ParseUint(assignment2Id, 10, 64)
	assignment2Hash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment2Instance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignment2Id,
			"docId_i":     assignment2IdI,
			"hash":        assignment2Hash,
			"createdDate": "2020-11-15T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "bob",
			"votes":       30,
		},
	)
	err = client.Mutate(assignment2Instance.AddMutation(false))
	assert.NilError(t, err)

	//***Add a new assignment to edge
	assignmentRef = map[string]interface{}{
		"docId": assignment2Id,
	}
	setPersonValues = map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef},
	}
	mutation, err = personType.UpdateMutation("docId", personId, setPersonValues, nil)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err = client.GetOne("docId", personId, personType, nil)
	assert.NilError(t, err)
	allPersonValues["assignments"] = append(allPersonValues["assignments"].([]map[string]interface{}), assignmentRef)
	personInstance = gql.NewSimplifiedInstance(
		personType,
		allPersonValues,
	)

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

	//***Remove an assignment from edge

	assignmentRef = map[string]interface{}{
		"docId": assignment1Id,
	}
	removePersonValues := map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef},
	}

	mutation, err = personType.UpdateMutation("docId", personId, nil, removePersonValues)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err = client.GetOne("docId", personId, personType, nil)
	assert.NilError(t, err)

	allPersonValues["assignments"] = allPersonValues["assignments"].([]map[string]interface{})[1:]

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

}

func TestUpdateEdgeToMoreGenericType(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType1 := gql.NewSimplifiedType(
		"Assignment1",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentType2 := gql.NewSimplifiedType(
		"Assignment2",
		map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
		},
		gql.DocumentSimplifiedInterface,
	)

	personType := gql.NewSimplifiedType(
		"Person",
		map[string]*gql.SimplifiedField{
			"name": {
				Name:    "name",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
		},
		gql.DocumentSimplifiedInterface,
	)

	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Hash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment1Instance := gql.NewSimplifiedInstance(
		assignmentType1,
		map[string]interface{}{
			"docId":       assignment1Id,
			"docId_i":     assignment1IdI,
			"hash":        assignment1Hash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment1",
			"assignee":    "alice",
			"votes":       20,
		},
	)

	assignment2Id := "2"
	assignment2IdI, _ := strconv.ParseUint(assignment2Id, 10, 64)
	assignment2Hash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951f"
	assignment2Instance := gql.NewSimplifiedInstance(
		assignmentType2,
		map[string]interface{}{
			"docId":       assignment2Id,
			"docId_i":     assignment2IdI,
			"hash":        assignment2Hash,
			"createdDate": "2020-10-12T18:27:47.000Z",
			"creator":     "dao.hypha1",
			"type":        "assignment2",
			"name":        "assign1",
		},
	)

	personId := "3"
	personIdI, _ := strconv.ParseUint(personId, 10, 64)
	personHash := "f4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	allPersonValues := map[string]interface{}{
		"docId":       personId,
		"docId_i":     personIdI,
		"hash":        personHash,
		"createdDate": "2020-11-12T18:27:47.000Z",
		"creator":     "dao.hypha",
		"type":        "person",
		"name":        "alice",
	}
	personInstance := gql.NewSimplifiedInstance(
		personType,
		allPersonValues,
	)

	updateOp, err := schema.UpdateType(assignmentType1)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)

	updateOp, err = schema.UpdateType(assignmentType2)
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
	err = client.Mutate(assignment2Instance.AddMutation(false))
	assert.NilError(t, err)
	err = client.Mutate(personInstance.AddMutation(false))
	assert.NilError(t, err)

	personType.Fields["assignments"] = gql.NewEdgeField("assignments", "Assignment1")
	assignmentRef1 := map[string]interface{}{
		"docId": assignment1Id,
	}
	setPersonValues := map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef1},
	}
	// oldPerson, err = schema.GetSimplifiedType("Person")
	// assert.NilError(t, err)
	// if reflect.ValueOf(oldPerson.Fields).Pointer() == reflect.ValueOf(personType.Fields).Pointer() {
	// 	fmt.Println("Same before update type 1")
	// } else {
	// 	fmt.Println("Different before update type 1")
	// }
	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	_, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	mutation, err := personType.UpdateMutation("docId", personId, setPersonValues, nil)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err := client.GetOne("docId", personId, personType, nil)
	assert.NilError(t, err)

	allPersonValues["assignments"] = []map[string]interface{}{assignmentRef1}
	personInstance = gql.NewSimplifiedInstance(
		personType,
		allPersonValues,
	)

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)
	personType.Fields["assignments"] = gql.NewEdgeField("assignments", "Document")
	assignmentRef2 := map[string]interface{}{
		"docId": assignment2Id,
	}
	setPersonValues = map[string]interface{}{
		"assignments": []map[string]interface{}{assignmentRef2},
	}
	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	// fmt.Println("Schema: ", schema.String())
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	_, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	mutation, err = personType.UpdateMutation("docId", personId, setPersonValues, nil)
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualPersonInstance, err = client.GetOne("docId", personId, personType, nil)
	assert.NilError(t, err)

	allPersonValues["assignments"] = []map[string]interface{}{assignmentRef1, assignmentRef2}
	personInstance = gql.NewSimplifiedInstance(
		personType,
		allPersonValues,
	)

	// fmt.Println("Actual Person Instance: ", actualPersonInstance)
	assertInstance(t, actualPersonInstance, personInstance)

	//Shouldn't update for no changes
	updateOp, err = schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_None)

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
	expectedHash := expectedRef["docId"].(string)
	for _, actualRef := range actualRefs {
		if actualRef.(map[string]interface{})["docId"].(string) == expectedHash {
			return
		}
	}
	assert.Assert(t, false, fmt.Sprintf("Expected docId: %v not found", expectedHash))
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
	assignmentType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)
	assignmentId := "1"
	assignmentIdI, _ := strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	err = client.Mutate(assignmentInstance.AddMutation(false))
	assert.NilError(t, err)

	actualAssignmentInstance, err := client.GetOne("docId", assignmentId, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	mutation, err := assignmentInstance.DeleteMutation("docId")
	assert.NilError(t, err)
	err = client.Mutate(mutation)
	assert.NilError(t, err)

	actualAssignmentInstance, err = client.GetOne("docId", assignmentId, assignmentType, nil)
	assert.NilError(t, err)
	assert.Assert(t, actualAssignmentInstance == nil)

}

func TestMultipleMutations(t *testing.T) {
	beforeEach()
	schema, err := gql.NewSchema("", true)
	assert.NilError(t, err)
	assignmentType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
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
		gql.DocumentSimplifiedInterface,
	)

	assignmentId := "1"
	assignmentIdI, _ := strconv.ParseUint(assignmentId, 10, 64)
	assignmentHash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignmentInstance := gql.NewSimplifiedInstance(
		assignmentType,
		map[string]interface{}{
			"docId":       assignmentId,
			"docId_i":     assignmentIdI,
			"hash":        assignmentHash,
			"createdDate": "2020-11-12T18:27:47.000Z",
			"creator":     "dao.hypha",
			"type":        "assignment",
			"assignee":    "alice",
			"votes":       20,
		},
	)
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

	actualAssignmentInstance, err := client.GetOne("docId", assignmentId, assignmentType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualAssignmentInstance, assignmentInstance)

	actualCursorInstance, err := client.GetOne("id", cursorId, gql.CursorSimplifiedType, nil)
	assert.NilError(t, err)

	// fmt.Println("Actual Instance: ", actualAssignmentInstance)
	assertInstance(t, actualCursorInstance, cursorInstance)

}
