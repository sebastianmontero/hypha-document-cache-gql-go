package gql_test

import (
	"log"
	"testing"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/vektah/gqlparser/ast"
	"gotest.tools/assert"
)

var dg *dgraph.Dgraph
var admin *gql.Admin

func adminTestSetup() {
	if admin == nil {
		admin = gql.NewAdmin("http://localhost:8080/admin")
	}
	if dg == nil {
		var err error
		dg, err = dgraph.New("")
		if err != nil {
			log.Fatal(err, "Unable to create dgraph")
		}
	}
	err := dg.DropAll()
	if err != nil {
		log.Fatal(err, "Unable to drop all")
	}
}

func TestUpdateSchema(t *testing.T) {
	adminTestSetup()
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
	assignmentFields := []*gql.SimplifiedField{
		{
			Name:    "assignee",
			Type:    "String",
			NonNull: true,
			Index:   "term",
		},
		{
			Name:    "role",
			Type:    "Role",
			IsArray: true,
		},
	}
	err = schema.UpdateType("Assignment", assignmentFields, true)
	assert.NilError(t, err)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	assert.NilError(t, err)
	assert.Assert(t, currentSchema.GetType("Person") != nil)
	assert.Assert(t, currentSchema.GetType("Role") != nil)
	assertType(t, "Assignment", assignmentFields, true, currentSchema)

}

func TestUpdateType(t *testing.T) {
	adminTestSetup()
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

	newFields := []*gql.SimplifiedField{
		{
			Name:    "name",
			Type:    "String",
			Index:   "term",
			NonNull: false,
		},
		{
			Name:    "picks",
			Type:    "Int64",
			IsArray: true,
			NonNull: false,
		},
		{
			Name:    "age",
			Type:    "Int64",
			NonNull: false,
		},
	}

	// ***Adding Document interface
	err = schema.UpdateType("Person", newFields, false)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, "Person", newFields, false, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
}

func TestAddFieldIfNotExists(t *testing.T) {
	adminTestSetup()
	schemaDef :=
		`
			type Person {
				name: String! @search(by: [term])
				picks: [Int64!]!
			}
		`
	fields := []*gql.SimplifiedField{
		{
			Name:    "name",
			Type:    "String",
			Index:   "term",
			NonNull: true,
		},
		{
			Name:    "picks",
			Type:    "Int64",
			IsArray: true,
			NonNull: true,
		},
	}
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	assertType(t, "Person", fields, false, currentSchema)
	// fmt.Println("Schema: ", currentSchema)
	newField := &gql.SimplifiedField{
		Name:    "age",
		Type:    "Int64",
		NonNull: false,
	}
	added, err := schema.AddFieldIfNotExists("Person", newField)
	assert.NilError(t, err)
	assert.Equal(t, added, true)
	fields = append(fields, newField)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, "Person", fields, false, currentSchema)

	added, err = schema.AddFieldIfNotExists("Person", newField)
	assert.NilError(t, err)
	assert.Equal(t, added, false)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, "Person", fields, false, currentSchema)

	// fmt.Println("Schema: ", currentSchema)
}

func TestAddEdge(t *testing.T) {
	adminTestSetup()
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
	fields := []*gql.SimplifiedField{
		{
			Name:    "name",
			Type:    "String",
			Index:   "term",
			NonNull: true,
		},
		{
			Name:    "picks",
			Type:    "Int64",
			IsArray: true,
			NonNull: true,
		},
	}
	schema, err := gql.NewSchema(schemaDef, true)
	assert.NilError(t, err)
	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err := admin.GetCurrentSchema()
	assert.NilError(t, err)
	assertType(t, "Person", fields, false, currentSchema)
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
	fields = append(fields, newField)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, "Person", fields, false, currentSchema)

	added, err = schema.AddEdge("Person", "roles", "Role")
	assert.NilError(t, err)
	assert.Equal(t, added, false)

	err = admin.UpdateSchema(schema)
	assert.NilError(t, err)
	currentSchema, err = admin.GetCurrentSchema()
	// fmt.Println(gql.DefinitionToString(schema.GetType("Person"), 0))
	assert.NilError(t, err)
	assertType(t, "Person", fields, false, currentSchema)

	// fmt.Println("Schema: ", currentSchema)
}

func TestUpdateTypeShouldFailForInvalidUpdate(t *testing.T) {
	adminTestSetup()
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
	personFields := []*gql.SimplifiedField{
		{
			Name:  "name",
			Type:  "String",
			Index: "term",
		},
		{
			Name:    "picks",
			Type:    "Int64",
			IsArray: true,
		},
	}

	roleFields := []*gql.SimplifiedField{
		{
			Name:    "hash",
			Type:    "String",
			Index:   "exact",
			NonNull: true,
		},
		{
			Name:    "type",
			Type:    "String",
			Index:   "exact",
			NonNull: true,
		},
	}
	// ***Adding Document interface
	err = schema.UpdateType("Person", personFields, true)
	assert.ErrorContains(t, err, "can't add Document interface")

	// ***Removing Document interface
	err = schema.UpdateType("Role", roleFields, false)
	assert.ErrorContains(t, err, "can't remove Document interface to type")

	// ***Add non null field
	roleFields = append(roleFields, &gql.SimplifiedField{
		Name:    "name",
		Type:    "String",
		Index:   "term",
		NonNull: true,
	})
	err = schema.UpdateType("Role", roleFields, true)
	assert.ErrorContains(t, err, "can't add non null field")

	// ***From null to non null
	personFields[0].NonNull = true
	err = schema.UpdateType("Person", personFields, false)
	assert.ErrorContains(t, err, "can't make nullable field: name, not nullable")

	// ***From non array to array
	personFields[0].NonNull = false
	personFields[0].IsArray = true
	err = schema.UpdateType("Person", personFields, false)
	assert.ErrorContains(t, err, "can't make scalar field: name an array")

	// ***From array to non array
	personFields[0].IsArray = false
	personFields[1].IsArray = false
	err = schema.UpdateType("Person", personFields, false)
	assert.ErrorContains(t, err, "can't make array field: picks a scalar")

	// ***Change array type
	personFields[1].IsArray = true
	personFields[1].Type = "String"
	err = schema.UpdateType("Person", personFields, false)
	assert.ErrorContains(t, err, "can't make array field: picks of type: Int64, an array of type: String")

	// ***Change scalar type
	personFields[1].Type = "Int64"
	personFields[0].Type = "DateTime"
	err = schema.UpdateType("Person", personFields, false)
	assert.ErrorContains(t, err, "can't make scalar field: name of type: String, a scalar of type: DateTime")

}

func assertType(t *testing.T, name string, fields []*gql.SimplifiedField, extendsDocument bool, schema *gql.Schema) {
	typeDef := schema.GetType(name)
	assert.Assert(t, typeDef != nil)
	if extendsDocument {
		assert.Equal(t, gql.HasInterface(typeDef, "Document"), true)
		assert.Equal(t, len(fields)+len(gql.DocumentFieldArgs), len(typeDef.Fields))
	} else {
		assert.Equal(t, gql.HasInterface(typeDef, "Document"), false)
		assert.Equal(t, len(fields), len(typeDef.Fields))
	}
	for _, field := range fields {
		fieldDef := typeDef.Fields.ForName(field.Name)
		assert.Assert(t, fieldDef != nil)
		assertField(t, field, fieldDef)
	}
}

func assertField(t *testing.T, expected *gql.SimplifiedField, actual *ast.FieldDefinition) {
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.NonNull, actual.Type.NonNull)
	if expected.IsArray {
		assert.Assert(t, actual.Type.Elem != nil)
		assert.Equal(t, expected.Type, actual.Type.Elem.NamedType)
	} else {
		assert.Assert(t, actual.Type.Elem == nil)
		assert.Equal(t, expected.Type, actual.Type.NamedType)
	}
	if expected.Index != "" {
		directive := actual.Directives.ForName("search")
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
