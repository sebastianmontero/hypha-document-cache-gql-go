package util

import (
	"fmt"
	"sort"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/vektah/gqlparser/ast"
	"gotest.tools/assert"
)

func AssertSimplifiedInstance(t *testing.T, actual, expected *gql.SimplifiedInstance) {
	AssertSimplifiedType(t, actual.SimplifiedType, expected.SimplifiedType)
	// fmt.Println("actual:", actual)
	// fmt.Println("expected:", expected)
	assert.Equal(t, len(actual.Values), len(expected.Values))

	for name := range expected.Values {
		field := expected.SimplifiedType.GetField(name)
		assert.Assert(t, field != nil, "Expected field: %v not found, type: %v", name, expected.SimplifiedType)
		if field.IsEdge() {
			AssertEdge(t, actual.GetValue(name), expected.GetValue(name))
		} else if field.IsCoreEdge() {
			AssertCoreEdge(t, actual.GetValue(name), expected.GetValue(name))
		} else {
			assert.DeepEqual(t, actual.GetValue(name), expected.GetValue(name))
		}
	}
}

func AssertEdge(t *testing.T, actual, expected interface{}) {
	if expected == nil {
		assert.Assert(t, actual == nil, "Expected edge is nil, but actual is not: %v of type: %T", actual, actual)
	}
	a := actual.([]interface{})
	e := expected.([]map[string]interface{})
	assert.Equal(t, len(a), len(e))
	for _, expectedEdge := range e {
		AssertContainsEdge(t, expectedEdge, a)
	}
}

func AssertCoreEdge(t *testing.T, actual, expected interface{}) {
	if expected == nil {
		assert.Assert(t, actual == nil)
	} else {
		a := actual.(map[string]interface{})
		e := expected.(map[string]interface{})
		assert.Equal(t, a["docId"], e["docId"])
	}
}
func AssertContainsEdge(t *testing.T, edge map[string]interface{}, edges []interface{}) {
	for _, e := range edges {
		if e.(map[string]interface{})["docId"] == edge["docId"] {
			return
		}
	}
	assert.Assert(t, false, fmt.Sprintf("edge: %v, not found", edge))
}

func AssertSimplifiedBaseType(t *testing.T, actual, expected *gql.SimplifiedBaseType) {
	assert.Equal(t, actual.Name, expected.Name)
	assert.Equal(t, len(actual.Fields), len(expected.Fields), "Different number of fields actual: %v expected: %v", actual.Fields, expected.Fields)
	for name, field := range expected.Fields {
		AssertSimplifiedField(t, actual.Fields[name], field)
	}
}

func AssertSimplifiedInterface(t *testing.T, actual, expected *gql.SimplifiedInterface) {
	AssertSimplifiedBaseType(t, actual.SimplifiedBaseType, expected.SimplifiedBaseType)
	AssertUnorderedStrArray(t, actual.SignatureFields, expected.SignatureFields)
	assert.Equal(t, len(actual.Types), len(expected.Types))
	for et, _ := range expected.Types {
		_, ok := actual.Types[et]
		assert.Assert(t, ok, "Expected type: %v not found in actual types: %v for interface:  %v", et, actual.Types, expected.Name)
	}
}

func AssertSimplifiedType(t *testing.T, actual, expected *gql.SimplifiedType) {
	AssertSimplifiedBaseType(t, actual.SimplifiedBaseType, expected.SimplifiedBaseType)
	AssertUnorderedStrArray(t, actual.Interfaces, expected.Interfaces)
}

func AssertSimplifiedField(t *testing.T, actual, expected *gql.SimplifiedField) {
	assert.Assert(t, actual != nil, fmt.Sprintf("For field '%v'", expected.Name))
	assert.Equal(t, actual.IsID, expected.IsID)
	assert.Equal(t, actual.Name, expected.Name)
	assert.Equal(t, actual.NonNull, expected.NonNull)
	assert.Equal(t, actual.Index, expected.Index)
	assert.Equal(t, actual.IsArray, expected.IsArray)
	assert.Equal(t, actual.Type, expected.Type)
}

func AssertSimplifiedInterfaces(t *testing.T, actual, expected gql.SimplifiedInterfaces) {
	assert.Equal(t, len(actual), len(expected), "Different number of interfaces actual: %v, expected: %v", len(actual), len(expected))
	for eName, eInterf := range expected {
		aInterf, ok := actual[eName]
		assert.Assert(t, ok, "Actual does not contain expected interface: %v", eName)
		AssertSimplifiedInterface(t, aInterf, eInterf)
	}
}

func AssertUnorderedStrArray(t *testing.T, actual, expected []string) {
	if len(expected) == 0 {
		assert.Equal(t, len(actual), 0)
		return
	}
	sort.Strings(actual)
	sort.Strings(expected)
	assert.DeepEqual(t, actual, expected)
}

func AssertType(t *testing.T, expected *gql.SimplifiedType, schema *gql.Schema) {

	typeDef := AssertBaseType(t, ast.Object, expected.SimplifiedBaseType, schema)
	AssertUnorderedStrArray(t, typeDef.Interfaces, expected.Interfaces)
}

func AssertInterface(t *testing.T, expected *gql.SimplifiedInterface, schema *gql.Schema) {
	AssertBaseType(t, ast.Interface, expected.SimplifiedBaseType, schema)
}

func AssertBaseType(t *testing.T, kind ast.DefinitionKind, expected *gql.SimplifiedBaseType, schema *gql.Schema) *ast.Definition {
	typeDef := schema.GetType(expected.Name)
	assert.Assert(t, typeDef != nil, fmt.Sprintf("For type: %v", expected.Name))
	assert.Equal(t, kind, typeDef.Kind)
	assert.Equal(t, len(expected.Fields), len(typeDef.Fields))
	for _, field := range expected.Fields {
		fieldDef := typeDef.Fields.ForName(field.Name)
		assert.Assert(t, fieldDef != nil, "Expected field %v definition for type: %v not found", field.Name, expected.Name)
		AssertField(t, field, fieldDef)
	}
	return typeDef
}

func AssertField(t *testing.T, expected *gql.SimplifiedField, actual *ast.FieldDefinition) {
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
