package gql_test

import (
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

func TestUpdateSchema(t *testing.T) {
	// schemaDef := "type Person { name: String }"
	schemaDef :=
		`
			type Role implements Document { ` +
			gql.DocumentFields + `
				name: String
			}
		`
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	// fmt.Println("D Schema: ", schema.String())
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	// fmt.Println("Schema: ", currentSchema)
	// fmt.Println(gql.DefinitionToString(currentSchema.GetType("Document"), 0))
	// fmt.Println(gql.DefinitionToString(currentSchema.GetType("Role"), 0))
	assert.Assert(t, currentSchema.GetType("Role") != nil)

	//***Add Type using string***
	schemaDef +=
		`
			type Person implements Document {` +
			gql.DocumentFields + `
				name: String @search(by: [term])
				createdAt: DateTime @search(by: [day])
				intValue: Int64 @search(by: [int64])
				picks: [Int64]
				role:[Role]
				role2: Role
			}
		`
	schema, err = gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	person := currentSchema.GetType("Person")
	assert.Assert(t, person != nil)
	assert.Assert(t, currentSchema.GetType("Role") != nil)
	// fmt.Println(gql.DefinitionToString(person, 0))
	// fmt.Println("Schema: ", currentSchema)

	//***Add Type programatically***
	assignmentType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
			"assignee": {
				Name:    "assignee",
				Type:    "String",
				NonNull: true,
				Index:   "term",
			},
			"role": {
				Name:    "role",
				Type:    "Role",
				IsArray: true,
			},
		},
		gql.DocumentSimplifiedInterface,
	)

	assert.NilError(t, err)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	assert.Assert(t, currentSchema.GetType("Person") != nil)
	assert.Assert(t, currentSchema.GetType("Role") != nil)
	util.AssertType(t, assignmentType, currentSchema)
	// fmt.Println("Schema: ", currentSchema.String())
	// *** There shouldn't be any changes for updating schema with the same type
	updateOp, err = schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_None)

	updateOp, err = currentSchema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_None)

	//***Add Type programatically with id***
	badgeType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Badge",
			Fields: map[string]*gql.SimplifiedField{
				"name": {
					IsID:    true,
					Name:    "name",
					Type:    "String",
					NonNull: true,
					Index:   "term",
				},
			},
		},
	}
	updateOp, err = schema.UpdateType(badgeType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	// for i := 0; i < 2000; i++ {
	// 	tType := &gql.SimplifiedType{
	// 		Name: fmt.Sprintf("Assignment%v", i),
	// 		Fields: map[string]*gql.SimplifiedField{
	// 			"assignee": {
	// 				Name:    "assignee",
	// 				Type:    "String",
	// 				NonNull: true,
	// 				Index:   "term",
	// 			},
	// 			"role": {
	// 				Name:    "role",
	// 				Type:    "Role",
	// 				IsArray: true,
	// 			},
	// 		},
	// 		ExtendsDocument: true,
	// 	}
	// 	updateOp, err = schema.UpdateType(tType)
	// 	assert.NilError(t, err)
	// 	assert.Equal(t, updateOp, true)
	// }

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	// fmt.Println("Schema: ", currentSchema.String())
	assert.Assert(t, currentSchema.GetType("Person") != nil)
	assert.Assert(t, currentSchema.GetType("Role") != nil)
	util.AssertType(t, assignmentType, currentSchema)
	util.AssertType(t, badgeType, currentSchema)
}

func TestSetInterface(t *testing.T) {

	schema, err := gql.NewSchema("", false)
	assert.NilError(t, err)
	schema.SetInterface(gql.DocumentSimplifiedInterface)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	util.AssertInterface(t, gql.DocumentSimplifiedInterface, currentSchema)

}

func TestAddTypeWithMultipleInterfaces(t *testing.T) {

	schema, err := gql.NewSchema("", false)
	assert.NilError(t, err)
	schema.SetInterface(gql.DocumentSimplifiedInterface)
	votableInterface := GetVotableInterface()
	schema.SetInterface(votableInterface)
	// fmt.Println(gql.DefinitionToString(person, 0))
	// fmt.Println("Schema: ", currentSchema)

	//***Add Type programatically***
	assignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Assignment",
			Fields: map[string]*gql.SimplifiedField{
				"assignee": {
					Name:    "assignee",
					Type:    "String",
					NonNull: true,
					Index:   "term",
				},
			},
		},
		Interfaces: []string{"Document", "Votable"},
	}

	assignmentType.SetFields(gql.DocumentFieldArgs)
	assignmentType.SetFields(votableInterface.Fields)
	updateOp, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	util.AssertInterface(t, gql.DocumentSimplifiedInterface, currentSchema)
	util.AssertInterface(t, votableInterface, currentSchema)
	util.AssertType(t, assignmentType, currentSchema)
}

func TestUpdateType(t *testing.T) {
	schemaDef :=
		`
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	assert.Assert(t, currentSchema.GetType("Person") != nil)
	// fmt.Println("Schema: ", currentSchema)

	personType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Person",
			Fields: map[string]*gql.SimplifiedField{
				"name": {
					Name:    "name",
					Type:    "String",
					Index:   "term",
					NonNull: false,
				},
				"picks": {
					Name:    "picks",
					Type:    "Int64",
					IsArray: true,
					NonNull: false,
				},
				"age": {
					Name:    "age",
					Type:    "Int64",
					NonNull: false,
				},
			},
		},
	}

	// ***Adding age field
	updateOp, err := schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)
	// // fmt.Println("Schema: ", currentSchema)
	// // ***Shouldn't change for same type
	// updateOp, err = currentSchema.UpdateType(personType)
	// assert.NilError(t, err)
	// assert.Equal(t, updateOp, gql.SchemaUpdateOp_None)

	// //***Removing field
	// removedField := personType.Fields["picks"]
	// delete(personType.Fields, "picks")
	// // fmt.Println("Person: ", personType)
	// updateOp, err = schema.UpdateType(personType)
	// assert.NilError(t, err)
	// assert.Equal(t, updateOp, gql.SchemaUpdateOp_None)

	// //***Adding field
	// personType.Fields["hobbie"] = &gql.SimplifiedField{
	// 	Name:    "hobbie",
	// 	Type:    "String",
	// 	NonNull: false,
	// }
	// // fmt.Println("Person: ", personType)
	// updateOp, err = schema.UpdateType(personType)
	// assert.NilError(t, err)
	// assert.Equal(t, updateOp, gql.SchemaUpdateOp_Updated)

	// err = admin.UpdateSchema(schema)
	// assert.NilError(t, err)
	// currentSchema, err = admin.GetCurrentSchema()
	// // fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	// assert.NilError(t, err)
	// //***Adding removed field, underlying schema should still have it
	// personType.Fields["picks"] = removedField
	// // ***Check simplified type was updated correctly
	// actualPersonType, err := schema.GetSimplifiedType("Person")
	// assert.NilError(t, err)
	// fmt.Println("actual: ", actualPersonType)
	// fmt.Println("expected: ", personType)
	// util.AssertSimplifiedType(t, actualPersonType, personType)
	// util.AssertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)

}

func TestUpdateField(t *testing.T) {
	schemaDef :=
		`
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`
	personType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Person",
			Fields: map[string]*gql.SimplifiedField{
				"name": {
					Name:    "name",
					Type:    "String",
					Index:   "term",
					NonNull: true,
				},
				"picks": {
					Name:    "picks",
					Type:    "Int64",
					IsArray: true,
					NonNull: true,
				},
			},
		},
	}

	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	// fmt.Println("Schema: ", schema)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	newField := &gql.SimplifiedField{
		Name:    "age",
		Type:    "Int64",
		NonNull: false,
	}
	updated, err := schema.UpdateField("Person", newField)
	assert.NilError(t, err)
	assert.Equal(t, updated, true)
	personType.Fields["age"] = newField
	// ***Check simplified type was updated correctly
	actualPersonType, err := schema.GetSimplifiedType("Person")
	assert.NilError(t, err)
	util.AssertSimplifiedType(t, actualPersonType, personType)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	// fmt.Println("Schema: ", currentSchema.String())
	util.AssertType(t, personType, currentSchema)

	updated, err = schema.UpdateField("Person", newField)
	assert.NilError(t, err)
	assert.Equal(t, updated, false)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)

	updateField := &gql.SimplifiedField{
		Name:    "name",
		Type:    "String",
		Index:   "term",
		NonNull: false,
	}
	updated, err = schema.UpdateField("Person", updateField)
	assert.NilError(t, err)
	assert.Equal(t, updated, true)
	personType.Fields["name"].NonNull = false
	// ***Check simplified type was updated correctly
	actualPersonType, err = schema.GetSimplifiedType("Person")
	assert.NilError(t, err)
	util.AssertSimplifiedType(t, actualPersonType, personType)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	// fmt.Println("Schema: ", currentSchema.String())
	util.AssertType(t, personType, currentSchema)

	// fmt.Println("Schema: ", currentSchema)
}

func TestUpdateFieldShouldFailForNonNullField(t *testing.T) {
	schemaDef :=
		`
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`

	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)

	newField := &gql.SimplifiedField{
		Name:    "age",
		Type:    "Int64",
		NonNull: true,
	}
	_, err = schema.UpdateField("Person", newField)
	assert.ErrorContains(t, err, "can't add non null field")
}

func TestUpdateEdge(t *testing.T) {
	schemaDef :=
		`
			type Role implements Document { ` +
			gql.DocumentFields + `
				name: String! @search(by: [term])
			}
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`
	personType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Person",
			Fields: map[string]*gql.SimplifiedField{
				"name": {
					Name:    "name",
					Type:    "String",
					Index:   "term",
					NonNull: true,
				},
				"picks": {
					Name:    "picks",
					Type:    "Int64",
					IsArray: true,
					NonNull: true,
				},
			},
		},
	}
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	newField := &gql.SimplifiedField{
		Name:    "roles",
		Type:    "Role",
		NonNull: false,
		IsArray: true,
	}

	updated, err := schema.UpdateEdge("Person", "roles", "Role")
	assert.NilError(t, err)
	assert.Equal(t, updated, true)
	personType.Fields["roles"] = newField

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)

	updated, err = schema.UpdateEdge("Person", "roles", "Role")
	assert.NilError(t, err)
	assert.Equal(t, updated, false)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)

	newField = &gql.SimplifiedField{
		Name:    "roles",
		Type:    "Document",
		NonNull: false,
		IsArray: true,
	}

	updated, err = schema.UpdateEdge("Person", "roles", "Document")
	assert.NilError(t, err)
	assert.Equal(t, updated, true)
	personType.Fields["roles"] = newField

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)

	// fmt.Println("Schema: ", currentSchema)
}

func TestUpdateEdgeShouldFailForInvalidTypeChange(t *testing.T) {
	schemaDef :=
		`
			type Role implements Document { ` +
			gql.DocumentFields + `
				name: String! @search(by: [term])
			}
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
				roles: [Document]
			}
		`
	personType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Person",
			Fields: map[string]*gql.SimplifiedField{
				"name": {
					Name:    "name",
					Type:    "String",
					Index:   "term",
					NonNull: true,
				},
				"picks": {
					Name:    "picks",
					Type:    "Int64",
					IsArray: true,
					NonNull: true,
				},
				"roles": {
					Name:    "roles",
					Type:    "Document",
					NonNull: false,
					IsArray: true,
				},
			},
		},
	}
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	util.AssertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	_, err = schema.UpdateEdge("Person", "roles", "Role")
	assert.ErrorContains(t, err, "Person, error: can't make array field: roles of type: Document, array of type: Role")
}

func TestUpdateTypeShouldFailForInvalidUpdate(t *testing.T) {
	schemaDef :=
		`
			type Person {
				name: String @search(by: [term])
				picks: [Int64]
			}
			type Role implements Document {` +
			gql.DocumentFields + `
			}
		`
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)

	roleType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Role",
			Fields: map[string]*gql.SimplifiedField{
				"hash": {
					Name:    "hash",
					Type:    "String",
					Index:   "exact",
					NonNull: true,
				},
				"type": {
					Name:    "type",
					Type:    "String",
					Index:   "exact",
					NonNull: true,
				},
			},
		},
	}

	personType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Person",
			Fields: map[string]*gql.SimplifiedField{
				"name": {
					Name:  "name",
					Type:  "String",
					Index: "term",
				},
				"picks": {
					Name:    "picks",
					Type:    "Int64",
					IsArray: true,
				},
			},
		},
	}
	// personType.AddInterface(gql.DocumentSimplifiedInterface)
	// // ***Adding Document interface
	// _, err = schema.UpdateType(personType)
	// assert.ErrorContains(t, err, "can't add Document interface")

	// // ***Removing Document interface
	// _, err = schema.UpdateType(roleType)
	// assert.ErrorContains(t, err, "can't remove Document interface to type")

	// // ***Add non null field
	// personType.ExtendsDocument = false
	// roleType.ExtendsDocument = true
	roleType.Fields["name"] = &gql.SimplifiedField{
		Name:    "name",
		Type:    "String",
		Index:   "term",
		NonNull: true,
	}

	_, err = schema.UpdateType(roleType)
	assert.ErrorContains(t, err, "can't add non null field")

	// ***From null to non null
	// personType.Fields["name"].NonNull = true
	// _, err = schema.UpdateType(personType)
	// assert.ErrorContains(t, err, "can't make nullable field: name, not nullable")

	// ***From non array to array
	personType.Fields["name"].NonNull = false
	personType.Fields["name"].IsArray = true
	_, err = schema.UpdateType(personType)
	assert.ErrorContains(t, err, "can't make scalar field: name an array")

	// ***From array to non array
	personType.Fields["name"].IsArray = false
	personType.Fields["picks"].IsArray = false
	_, err = schema.UpdateType(personType)
	assert.ErrorContains(t, err, "can't make array field: picks a scalar")

	// ***Change array type
	personType.Fields["picks"].IsArray = true
	personType.Fields["picks"].Type = "String"
	_, err = schema.UpdateType(personType)
	assert.ErrorContains(t, err, "can't make array field: picks of type: Int64, array of type: String")

	// ***Change scalar type
	personType.Fields["picks"].Type = "Int64"
	personType.Fields["name"].Type = "DateTime"
	_, err = schema.UpdateType(personType)
	assert.ErrorContains(t, err, "can't make scalar field: name of type: String, scalar of type: DateTime")

}

// func TestInterfaceExtendsInterfaceSchema(t *testing.T) {
// 	// schemaDef := "type Person { name: String }"
// 	schemaDef :=
// 		`
// 			interface Role implements Document { ` +
// 			gql.DocumentFields + `
// 				name: String
// 			}
// 		`
// 	schema, err := gql.NewSchema(schemaDef, true)
// 	assert.NilError(t, err)
// 	// fmt.Println("D Schema: ", schema.String())
// 	err = admin.UpdateSchema(schema)
// 	assert.NilError(t, err)
// 	_, err = admin.GetCurrentSchema()
// 	assert.NilError(t, err)
// }

// func TestInterfaceWithMoreGenericTypeThanChild(t *testing.T) {
// 	// schemaDef := "type Person { name: String }"
// 	schemaDef :=
// 		`
// 			type Role implements Document { ` +
// 			gql.DocumentFields + `
// 				name: String
// 			}

// 			interface User {
// 				account: String
// 				role: [Document]
// 			}

// 			type Member implements User {
// 				account: String
// 				role: [Role]
// 			}
// 		`
// 	schema, err := gql.NewSchema(schemaDef, true)
// 	assert.NilError(t, err)
// 	// fmt.Println("D Schema: ", schema.String())
// 	err = admin.UpdateSchema(schema)
// 	assert.NilError(t, err)
// 	_, err = admin.GetCurrentSchema()
// 	assert.NilError(t, err)
// }

func GetVotableInterface() *gql.SimplifiedInterface {
	return gql.NewSimplifiedInterface(
		"Votable",
		map[string]*gql.SimplifiedField{
			"ballot_expiration_t": {
				Name:  "ballot_expiration_t",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
			"details_description_s": {
				Name:  "details_description_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
		},
		[]string{},
		nil,
	)
}
