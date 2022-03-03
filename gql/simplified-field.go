package gql

import (
	"fmt"
	"reflect"

	"github.com/vektah/gqlparser/ast"
)

type Indexes map[string]bool

func NewIndexes(indexes ...string) Indexes {
	idxs := make(Indexes, len(indexes))
	for _, index := range indexes {
		idxs[index] = true
	}
	return idxs
}

func NewIndexesFromFieldDef(fieldDef *ast.FieldDefinition) (Indexes, error) {
	directive := fieldDef.Directives.ForName("search")
	if directive != nil {
		argument := directive.Arguments.ForName("by")
		if argument == nil {
			return nil, fmt.Errorf("don't know how to parse index for type: %v", fieldDef.Name)
		}
		indexes := make(Indexes, len(argument.Value.Children))
		for _, child := range argument.Value.Children {
			indexes[child.Value.Raw] = true
		}
		return indexes, nil
	}
	return nil, nil
}

func (m Indexes) Has(index string) bool {
	_, ok := m[index]
	return ok
}

func (m Indexes) HasIndexes() bool {
	return len(m) > 0
}

func (m Indexes) Len() int {
	return len(m)
}

func (m Indexes) Equal(indexes Indexes) bool {
	return reflect.DeepEqual(m, indexes)
}

func (m Indexes) Clone() Indexes {
	idxs := make(Indexes, len(m))
	for k, v := range m {
		idxs[k] = v
	}
	return idxs
}

type SimplifiedField struct {
	IsID    bool
	Name    string
	Type    string
	NonNull bool
	Indexes Indexes
	IsArray bool
}

func NewSimplifiedField(fieldDef *ast.FieldDefinition) (*SimplifiedField, error) {
	field := &SimplifiedField{
		IsID:    fieldDef.Directives.ForName("id") != nil || fieldDef.Type.NamedType == GQLType_ID,
		Name:    fieldDef.Name,
		NonNull: fieldDef.Type.NonNull,
	}

	if fieldDef.Type.Elem == nil {
		field.Type = fieldDef.Type.NamedType
	} else {
		field.Type = fieldDef.Type.Elem.NamedType
		field.IsArray = true
	}

	indexes, err := NewIndexesFromFieldDef(fieldDef)
	if err != nil {
		return nil, err
	}
	field.Indexes = indexes
	return field, nil
}

func NewEdgeField(edgeName, edgeType string) *SimplifiedField {
	return &SimplifiedField{
		Name:    edgeName,
		Type:    edgeType,
		IsArray: true,
	}
}

func (m *SimplifiedField) Clone() *SimplifiedField {
	return &SimplifiedField{
		IsID:    m.IsID,
		Name:    m.Name,
		Type:    m.Type,
		NonNull: m.NonNull,
		Indexes: m.Indexes.Clone(),
		IsArray: m.IsArray,
	}
}

func (m *SimplifiedField) IsObject() bool {
	return m.Type != GQLType_Int64 && m.Type != GQLType_String && m.Type != GQLType_Time && m.Type != GQLType_ID
}

func (m *SimplifiedField) IsCoreEdge() bool {
	return !m.IsArray && m.IsObject()
}

func (m *SimplifiedField) IsEdge() bool {
	return m.IsArray && m.IsObject()
}

func (m *SimplifiedField) equal(field *SimplifiedField) bool {
	return m.IsID == field.IsID && m.Name == field.Name && m.Type == field.Type &&
		m.NonNull == field.NonNull && m.IsArray == field.IsArray && m.Indexes.Equal(field.Indexes)
}

func (m *SimplifiedField) CheckUpdate(new *SimplifiedField) error {

	// if new.IsID != m.IsID {
	// 	return fmt.Errorf("can't change id definition for field: %v", new.Name)
	// }
	// if new.NonNull && !m.NonNull {
	// 	return fmt.Errorf("can't make nullable field: %v, not nullable", new.Name)
	// }
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
		if m.IsObject() {
			if new.Type == DocumentSimplifiedInterface.Name {
				return nil
			}
		}
		// if new.Type == GQLType_Int64 && m.Type == GQLType_String {
		// 	return nil
		// }

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
				Indexes: %v,
			}		
		`,
		m.IsID,
		m.Name,
		m.Type,
		m.NonNull,
		m.IsArray,
		m.Indexes,
	)
}
