package gql

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/formatter"
)

type Schema struct {
	Schema          *ast.Schema
	SimplifiedTypes map[string]*SimplifiedType
}

func InitialSchema() (*Schema, error) {
	return NewSchema("", true)
}

func LoadSchema(schemaDef string) (*Schema, error) {
	return NewSchema(schemaDef, false)
}

func NewSchema(schemaDef string, includeBaseSchema bool) (*Schema, error) {
	sources := []*ast.Source{
		DgraphSchemaSource,
	}
	if includeBaseSchema {
		sources = append(sources, BaseSchemaSource)
	}
	if schemaDef != "" {
		sources = append(sources, &ast.Source{
			Input: schemaDef,
		})
	}
	schema, gqlErr := gqlparser.LoadSchema(sources...)
	if gqlErr != nil {
		return nil, fmt.Errorf("failed to parse schema, error: %v", gqlErr)
	}
	return &Schema{
		Schema:          schema,
		SimplifiedTypes: make(map[string]*SimplifiedType),
	}, nil
}

func (m *Schema) GetSimplifiedType(name string) (*SimplifiedType, error) {
	simplifiedType, ok := m.SimplifiedTypes[name]
	if !ok {
		typeDef := m.GetType(name)
		if typeDef == nil {
			return nil, nil
		}
		simplifiedType, err := NewSimplifiedType(typeDef)
		if err != nil {
			return nil, err
		}
		m.SimplifiedTypes[name] = simplifiedType
	}
	return simplifiedType, nil
}

func (m *Schema) GetType(name string) *ast.Definition {
	if typeDef, ok := m.Schema.Types[name]; ok {
		return typeDef
	}
	return nil
}

func (m *Schema) UpdateType(name string, fields []*SimplifiedField, extendsDocument bool) error {
	oldType, ok := m.Schema.Types[name]
	if !ok {
		m.Schema.Types[name] = CreateType(name, fields, extendsDocument)
		return nil
	}
	if extendsDocument && !HasInterface(oldType, "Document") {
		return fmt.Errorf("can't add Document interface to type: %v", name)
	}
	if !extendsDocument && HasInterface(oldType, "Document") {
		return fmt.Errorf("can't remove Document interface to type: %v", name)
	}
	fieldDefs := &oldType.Fields
	toAdd := make([]*SimplifiedField, 0)
	toUpdate := make([]*SimplifiedField, 0)
	for _, field := range fields {
		oldField := fieldDefs.ForName(field.Name)
		if oldField == nil {
			if field.NonNull {
				return fmt.Errorf("can't add non null field: %v to type: %v", field.Name, name)
			}
			toAdd = append(toAdd, field)
		} else {
			err := assertFieldUpdateIsValid(oldField, field)
			if err != nil {
				return err
			}
			toUpdate = append(toUpdate, field)
		}
	}
	for _, field := range toUpdate {
		pos := findFieldPos(field.Name, *fieldDefs)
		(*fieldDefs)[pos] = CreateField(field)
	}
	for _, field := range toAdd {
		*fieldDefs = append(*fieldDefs, CreateField(field))
	}
	return nil
}

func (m *Schema) AddEdge(typeName, edgeName, edgeType string) (bool, error) {
	return m.AddFieldIfNotExists(typeName, &SimplifiedField{
		Name:    edgeName,
		Type:    edgeType,
		IsArray: true,
	})
}

func (m *Schema) AddFieldIfNotExists(typeName string, field *SimplifiedField) (bool, error) {
	typeDef := m.GetType(typeName)
	if typeDef == nil {
		return false, fmt.Errorf("failed to add field, type: %v not found", typeName)
	}
	if fieldDef := typeDef.Fields.ForName(field.Name); fieldDef == nil {
		fieldDefs := &typeDef.Fields
		*fieldDefs = append(*fieldDefs, CreateField(field))
		return true, nil
	}
	return false, nil
}

func ExtendsDocument(typeDef *ast.Definition) bool {
	return HasInterface(typeDef, "Document")
}

func HasInterface(typeDef *ast.Definition, interfaceName string) bool {
	for _, intrfc := range typeDef.Interfaces {
		if intrfc == interfaceName {
			return true
		}
	}
	return false
}

func findFieldPos(name string, l ast.FieldList) int {
	for i, it := range l {
		if it.Name == name {
			return i
		}
	}
	return -1
}

func assertFieldUpdateIsValid(oldField *ast.FieldDefinition, newField *SimplifiedField) error {
	if newField.NonNull && !oldField.Type.NonNull {
		return fmt.Errorf("can't make nullable field: %v, not nullable", newField.Name)
	}
	if newField.IsArray && oldField.Type.Elem == nil {
		return fmt.Errorf("can't make scalar field: %v an array", newField.Name)
	}
	if !newField.IsArray && oldField.Type.Elem != nil {
		return fmt.Errorf("can't make array field: %v a scalar", newField.Name)
	}
	if newField.IsArray && newField.Type != oldField.Type.Elem.NamedType {
		return fmt.Errorf("can't make array field: %v of type: %v, an array of type: %v", newField.Name, oldField.Type.Elem.NamedType, newField.Type)
	}
	if !newField.IsArray && newField.Type != oldField.Type.NamedType {
		return fmt.Errorf("can't make scalar field: %v of type: %v, a scalar of type: %v", newField.Name, oldField.Type.NamedType, newField.Type)
	}
	return nil
}

func (m *Schema) String() string {
	out := &strings.Builder{}
	fmttr := formatter.NewFormatter(out)
	fmttr.FormatSchema(m.Schema)
	return out.String()
}

func CreateType(name string, fields []*SimplifiedField, extendsDocument bool) *ast.Definition {
	interfaces := []string{}
	var fieldDefs ast.FieldList
	if extendsDocument {
		fieldDefs = addFields(DocumentFieldArgs, fieldDefs)
		interfaces = append(interfaces, "Document")
	}
	fieldDefs = addFields(fields, fieldDefs)

	return &ast.Definition{
		Kind:       ast.Object,
		Name:       name,
		Fields:     fieldDefs,
		Interfaces: interfaces,
	}
}

func addFields(fields []*SimplifiedField, fieldList ast.FieldList) ast.FieldList {
	for _, field := range fields {
		fieldList = append(fieldList, CreateField(field))
	}
	return fieldList
}

func CreateField(args *SimplifiedField) *ast.FieldDefinition {

	fieldType := &ast.Type{
		NonNull: args.NonNull,
	}
	if args.IsArray {
		fieldType.Elem = &ast.Type{
			NamedType: args.Type,
			NonNull:   true,
		}
	} else {
		fieldType.NamedType = args.Type
	}
	var directives ast.DirectiveList

	if args.Index != "" {
		directives = ast.DirectiveList{
			{
				Name: "search",
				Arguments: ast.ArgumentList{
					{
						Name: "by",
						Value: &ast.Value{
							Kind: ast.ListValue,
							Children: ast.ChildValueList{
								{
									Value: &ast.Value{
										Raw:  args.Index,
										Kind: ast.EnumValue,
									},
								},
							},
						},
					},
				},
			},
		}
	}
	return &ast.FieldDefinition{
		Name:       args.Name,
		Type:       fieldType,
		Directives: directives,
	}
}

func DefinitionToString(def *ast.Definition, depth int) string {
	out := &strings.Builder{}
	out.WriteString(fmt.Sprintf("%vDefinition Name:%v\n", indent(depth), def.Name))
	out.WriteString(fmt.Sprintf("%vInterfaces: %v\n", indent(depth+1), def.Interfaces))
	out.WriteString(fmt.Sprintf("%vTypes: %v\n\n", indent(depth+1), def.Types))
	out.WriteString(FieldsToString(def.Fields, depth+1))
	return out.String()
}

func FieldsToString(fields ast.FieldList, depth int) string {
	out := &strings.Builder{}
	for _, field := range fields {
		out.WriteString(fmt.Sprintf("%vField Name:  %v\n", indent(depth), field.Name))
		out.WriteString(fmt.Sprintf("%vField Description:  %v\n", indent(depth), field.Description))
		out.WriteString(TypeToString(field.Type, depth))
		out.WriteString(DirectivesToString(field.Directives, depth))
		out.WriteString(ArgumentDefinitionListToString(field.Arguments, depth))
	}
	return out.String()
}

func ArgumentDefinitionListToString(arguments ast.ArgumentDefinitionList, depth int) string {
	out := &strings.Builder{}
	for _, argument := range arguments {
		out.WriteString(fmt.Sprintf("%vArgument Definition Name: %v\n", indent(depth), argument.Name))
		if argument.Type != nil {
			out.WriteString(TypeToString(argument.Type, depth))
		}
		if argument.DefaultValue != nil {
			out.WriteString(ValueToString(argument.DefaultValue, depth))
		}
		out.WriteString(DirectivesToString(argument.Directives, depth))
	}
	return out.String()
}

func DirectivesToString(directives ast.DirectiveList, depth int) string {
	out := &strings.Builder{}
	for _, directive := range directives {
		out.WriteString(fmt.Sprintf("%vDirective Name: %v\n", indent(depth), directive.Name))
		for _, argument := range directive.Arguments {
			out.WriteString(fmt.Sprintf("%vArgument Name: %v\n", indent(depth+1), argument.Name))
			out.WriteString(ValueToString(argument.Value, depth+4))
		}
	}
	return out.String()
}

func TypeToString(typeDef *ast.Type, depth int) string {
	out := &strings.Builder{}
	out.WriteString(fmt.Sprintf("%vField Type Named Type: %v, NonNull: %v, Position: %v\n", indent(depth), typeDef.NamedType, typeDef.NonNull, typeDef.Position))
	if typeDef.Elem != nil {
		out.WriteString(fmt.Sprintf("%vElem:\n", indent(depth)))
		out.WriteString(TypeToString(typeDef.Elem, depth+1))
	}
	return out.String()
}

func ValueToString(valueDef *ast.Value, depth int) string {
	out := &strings.Builder{}
	out.WriteString(fmt.Sprintf("%vField Value Raw: '%v', Kind: %v\n", indent(depth), valueDef.Raw, valueDef.Kind))
	if valueDef.Definition != nil {
		out.WriteString(DefinitionToString(valueDef.Definition, depth))
	}
	if valueDef.ExpectedType != nil {
		out.WriteString(TypeToString(valueDef.ExpectedType, depth))
	}
	if valueDef.VariableDefinition != nil {
		out.WriteString(VariableDefinitionToString(valueDef.VariableDefinition, depth))
	}
	for _, childValue := range valueDef.Children {
		out.WriteString(ChildValueToString(childValue, depth+1))
	}
	return out.String()
}

func VariableDefinitionToString(variableDef *ast.VariableDefinition, depth int) string {
	out := &strings.Builder{}
	out.WriteString(fmt.Sprintf("%vVariable Definition Variable: %v\n", indent(depth), variableDef.Variable))
	if variableDef.Definition != nil {
		out.WriteString(DefinitionToString(variableDef.Definition, depth))
	}
	if variableDef.Type != nil {
		out.WriteString(TypeToString(variableDef.Type, depth))
	}
	if variableDef.DefaultValue != nil {
		out.WriteString(ValueToString(variableDef.DefaultValue, depth))
	}
	return out.String()
}

func ChildValueToString(childValue *ast.ChildValue, depth int) string {
	out := &strings.Builder{}
	out.WriteString(fmt.Sprintf("%vChild Value: %v\n", indent(depth), childValue.Name))
	if childValue.Value != nil {
		out.WriteString(ValueToString(childValue.Value, depth))
	}
	return out.String()
}

func indent(depth int) string {
	return strings.Repeat("\t", depth)
}
