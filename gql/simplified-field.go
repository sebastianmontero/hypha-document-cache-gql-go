package gql

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

type SimplifiedField struct {
	IsID    bool
	Name    string
	Type    string
	NonNull bool
	Index   string
	IsArray bool
}

func NewSimplifiedField(fieldDef *ast.FieldDefinition) (*SimplifiedField, error) {
	field := &SimplifiedField{
		IsID:    fieldDef.Directives.ForName("id") != nil,
		Name:    fieldDef.Name,
		NonNull: fieldDef.Type.NonNull,
	}

	if fieldDef.Type.Elem == nil {
		field.Type = fieldDef.Type.NamedType
	} else {
		field.Type = fieldDef.Type.Elem.NamedType
		field.IsArray = true
	}

	directive := fieldDef.Directives.ForName("search")
	if directive != nil {
		argument := directive.Arguments.ForName("by")
		if argument == nil || len(argument.Value.Children) != 1 {
			return nil, fmt.Errorf("don't know how to parse index for type: %v", field.Name)
		}
		field.Index = argument.Value.Children[0].Value.Raw
	}

	return field, nil
}

func (m *SimplifiedField) CheckUpdate(new *SimplifiedField) error {

	if new.IsID != m.IsID {
		return fmt.Errorf("can't change id definition for field: %v", new.Name)
	}
	if new.NonNull && !m.NonNull {
		return fmt.Errorf("can't make nullable field: %v, not nullable", new.Name)
	}
	if new.IsArray && !m.IsArray {
		return fmt.Errorf("can't make scalar field: %v an array", new.Name)
	}
	if !new.IsArray && m.IsArray {
		return fmt.Errorf("can't make array field: %v a scalar", new.Name)
	}
	if new.Type != m.Type {
		cardinality := "scalar"
		if new.IsArray {
			cardinality = "array"
		}
		return fmt.Errorf("can't make %v field: %v of type: %v, %v of type: %v", cardinality, new.Name, m.Type, cardinality, new.Type)
	}
	return nil
}

func (m *SimplifiedField) String() string {
	return fmt.Sprintf(
		`
			SimplifiedField: {
				IsId: %v,
				Name: %v,
				Type: %v,
				NonNull: %v,
				IsArray: %v,
			}		
		`,
		m.IsID,
		m.Name,
		m.Type,
		m.NonNull,
		m.IsArray,
	)
}
