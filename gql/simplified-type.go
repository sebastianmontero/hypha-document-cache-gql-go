package gql

import (
	"fmt"

	"github.com/vektah/gqlparser/ast"
)

type SimplifiedType struct {
	Name            string
	Fields          map[string]*SimplifiedField
	ExtendsDocument bool
}

func NewSimplifiedType(typeDef *ast.Definition) (*SimplifiedType, error) {
	fields := make(map[string]*SimplifiedField)

	for _, fieldDef := range typeDef.Fields {
		field, err := NewFieldArgs(fieldDef)
		if err != nil {
			return nil, fmt.Errorf("failed to create simplified type from type definition for type: %v, error: %v", typeDef.Name, err)
		}
		fields[field.Name] = field
	}
	return &SimplifiedType{
		Name:            typeDef.Name,
		Fields:          fields,
		ExtendsDocument: ExtendsDocument(typeDef),
	}, nil
}
