package util

import (
	"fmt"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"gotest.tools/assert"
)

func AssertSimplifiedInstance(t *testing.T, actual, expected *gql.SimplifiedInstance) {
	AssertSimplifiedType(t, actual.SimplifiedType, expected.SimplifiedType)
	assert.Equal(t, len(actual.Values), len(expected.Values))

	for name := range expected.Values {
		field := expected.SimplifiedType.GetField(name)
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
		assert.Equal(t, a["hash"], e["hash"])
	}
}
func AssertContainsEdge(t *testing.T, edge map[string]interface{}, edges []interface{}) {
	for _, e := range edges {
		if e.(map[string]interface{})["hash"] == edge["hash"] {
			return
		}
	}
	assert.Assert(t, false, fmt.Sprintf("edge: %v, not found", edge))
}

func AssertSimplifiedType(t *testing.T, actual, expected *gql.SimplifiedType) {
	assert.Equal(t, actual.Name, expected.Name)
	assert.Equal(t, actual.ExtendsDocument, expected.ExtendsDocument)
	assert.Equal(t, len(actual.Fields), len(expected.Fields))
	for name, field := range expected.Fields {
		AssertSimplifiedField(t, actual.Fields[name], field)
	}
}

func AssertSimplifiedField(t *testing.T, actual, expected *gql.SimplifiedField) {
	assert.Assert(t, actual != nil, fmt.Sprintf("For field %v", expected.Name))
	assert.Equal(t, actual.IsID, expected.IsID)
	assert.Equal(t, actual.Name, expected.Name)
	assert.Equal(t, actual.NonNull, expected.NonNull)
	assert.Equal(t, actual.Index, expected.Index)
	assert.Equal(t, actual.IsArray, expected.IsArray)
	assert.Equal(t, actual.Type, expected.Type)
}
