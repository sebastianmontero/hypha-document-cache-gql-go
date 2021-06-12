package gql_test

import (
	"fmt"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/vektah/gqlparser/ast"
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
	assignmentType := &gql.SimplifiedType{
		Name: "Assignment",
		Fields: map[string]*gql.SimplifiedField{
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
		ExtendsDocument: true,
	}
	changed, err := schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, changed, true)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	assert.Assert(t, currentSchema.GetType("Person") != nil)
	assert.Assert(t, currentSchema.GetType("Role") != nil)
	assertType(t, assignmentType, currentSchema)
	// fmt.Println("Schema: ", currentSchema.String())
	// *** There shouldn't be any changes for updating schema with the same type
	changed, err = schema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, changed, false)

	changed, err = currentSchema.UpdateType(assignmentType)
	assert.NilError(t, err)
	assert.Equal(t, changed, false)

	//***Add Type programatically with id***
	badgeType := &gql.SimplifiedType{
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
		ExtendsDocument: false,
	}
	changed, err = schema.UpdateType(badgeType)
	assert.NilError(t, err)
	assert.Equal(t, changed, true)
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
	// 	changed, err = schema.UpdateType(tType)
	// 	assert.NilError(t, err)
	// 	assert.Equal(t, changed, true)
	// }

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	// fmt.Println("Schema: ", currentSchema.String())
	assert.Assert(t, currentSchema.GetType("Person") != nil)
	assert.Assert(t, currentSchema.GetType("Role") != nil)
	assertType(t, assignmentType, currentSchema)
	assertType(t, badgeType, currentSchema)
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
		ExtendsDocument: false,
	}

	// ***Adding Document interface
	changed, err := schema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, changed, true)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	// ***Shouldn't change for same type
	changed, err = currentSchema.UpdateType(personType)
	assert.NilError(t, err)
	assert.Equal(t, changed, false)
}

func TestAddFieldIfNotExists(t *testing.T) {
	schemaDef :=
		`
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`
	personType := &gql.SimplifiedType{
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
		ExtendsDocument: false,
	}

	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	// fmt.Println("Schema: ", schema)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	assertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	newField := &gql.SimplifiedField{
		Name:    "age",
		Type:    "Int64",
		NonNull: false,
	}
	added, err := schema.AddFieldIfNotExists("Person", newField)
	assert.NilError(t, err)
	assert.Equal(t, added, true)
	personType.Fields["age"] = newField

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	// fmt.Println("Schema: ", currentSchema.String())
	assertType(t, personType, currentSchema)

	added, err = schema.AddFieldIfNotExists("Person", newField)
	assert.NilError(t, err)
	assert.Equal(t, added, false)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, personType, currentSchema)

	// fmt.Println("Schema: ", currentSchema)
}

func TestAddFieldIfNotExistsShouldFailForNonNullField(t *testing.T) {
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
	_, err = schema.AddFieldIfNotExists("Person", newField)
	assert.ErrorContains(t, err, "can't add non null field")
}

func TestAddEdge(t *testing.T) {
	schemaDef :=
		`
			type Role {
				name: String! @search(by: [term])
			}
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`
	personType := &gql.SimplifiedType{
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
		ExtendsDocument: false,
	}
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	assertType(t, personType, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	newField := &gql.SimplifiedField{
		Name:    "roles",
		Type:    "Role",
		NonNull: false,
		IsArray: true,
	}

	added, err := schema.AddEdge("Person", "roles", "Role")
	assert.NilError(t, err)
	assert.Equal(t, added, true)
	personType.Fields["roles"] = newField

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, personType, currentSchema)

	added, err = schema.AddEdge("Person", "roles", "Role")
	assert.NilError(t, err)
	assert.Equal(t, added, false)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, personType, currentSchema)

	// fmt.Println("Schema: ", currentSchema)
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
		ExtendsDocument: false,
	}

	personType := &gql.SimplifiedType{
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
		ExtendsDocument: true,
	}
	// ***Adding Document interface
	_, err = schema.UpdateType(personType)
	assert.ErrorContains(t, err, "can't add Document interface")

	// ***Removing Document interface
	_, err = schema.UpdateType(roleType)
	assert.ErrorContains(t, err, "can't remove Document interface to type")

	// ***Add non null field
	personType.ExtendsDocument = false
	roleType.ExtendsDocument = true
	roleType.Fields["name"] = &gql.SimplifiedField{
		Name:    "name",
		Type:    "String",
		Index:   "term",
		NonNull: true,
	}

	_, err = schema.UpdateType(roleType)
	assert.ErrorContains(t, err, "can't add non null field")

	// ***From null to non null
	personType.Fields["name"].NonNull = true
	_, err = schema.UpdateType(personType)
	assert.ErrorContains(t, err, "can't make nullable field: name, not nullable")

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

func assertType(t *testing.T, expected *gql.SimplifiedType, schema *gql.Schema) {
	typeDef := schema.GetType(expected.Name)
	assert.Assert(t, typeDef != nil, fmt.Sprintf("For type: %v", expected.Name))
	if expected.ExtendsDocument {
		assert.Equal(t, gql.HasInterface(typeDef, "Document"), true)
		assert.Equal(t, len(expected.Fields)+len(gql.DocumentFieldArgs), len(typeDef.Fields))
	} else {
		assert.Equal(t, gql.HasInterface(typeDef, "Document"), false)
		assert.Equal(t, len(expected.Fields), len(typeDef.Fields))
	}
	for _, field := range expected.Fields {
		fieldDef := typeDef.Fields.ForName(field.Name)
		assert.Assert(t, fieldDef != nil)
		assertField(t, field, fieldDef)
	}
}

func assertField(t *testing.T, expected *gql.SimplifiedField, actual *ast.FieldDefinition) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.NonNull, actual.Type.NonNull, fmt.Sprintf("For field: %v", expected.Name))
	if expected.IsArray {
		assert.Assert(t, actual.Type.Elem != nil)
		assert.Equal(t, expected.Type, actual.Type.Elem.NamedType)
	} else {
		assert.Assert(t, actual.Type.Elem == nil)
		assert.Equal(t, expected.Type, actual.Type.NamedType)
	}
	directive := actual.Directives.ForName("id")
	if expected.IsID {
		assert.Assert(t, directive != nil)
	} else {
		assert.Assert(t, directive == nil)
	}
	if expected.Index != "" {
		directive = actual.Directives.ForName("search")
		assert.Assert(t, directive != nil)
		argument := directive.Arguments.ForName("by")
		assert.Assert(t, directive != nil)
		assert.Equal(t, ast.ListValue, argument.Value.Kind)
		value := argument.Value.Children[0].Value
		assert.Assert(t, value != nil)
		assert.Equal(t, expected.Index, value.Raw)
		assert.Equal(t, ast.EnumValue, value.Kind)
	} else {
		directive := actual.Directives.ForName("search")
		assert.Assert(t, directive == nil)
	}
}
