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
	for name, value := range expected.Values {
		assert.Equal(t, actual.Values[name], value)
	}
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
}
