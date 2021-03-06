package gql

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser"
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/formatter"
)

const (
	GQLType_ID     = "ID"
	GQLType_Int64  = "Int64"
	GQLType_Time   = "DateTime"
	GQLType_String = "String"
)

type SchemaUpdateOp string

const (
	SchemaUpdateOp_None    SchemaUpdateOp = "None"
	SchemaUpdateOp_Created SchemaUpdateOp = "Created"
	SchemaUpdateOp_Updated SchemaUpdateOp = "Updated"
)

// Provides the functionality for keeping track of the current schema state and updating it
type Schema struct {
	Schema          *ast.Schema
	SimplifiedTypes map[string]*SimplifiedType
}

func InitialSchema() (*Schema, error) {
	return NewSchema("", true)
}

// Initalizes the schema struct from the provided string schema
func LoadSchema(schemaDef string) (*Schema, error) {
	return NewSchema(schemaDef, false)
}

// Creates a new schema struct from the provided string schema, the
// includeBaseSchema parameter indicates whether the base schema should be added
// to the provided schema
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

// Returns the simplified type for the type with the specified name
func (m *Schema) GetSimplifiedType(name string) (*SimplifiedType, error) {
	simplifiedType, ok := m.SimplifiedTypes[name]
	if !ok {
		typeDef := m.GetType(name)
		if typeDef == nil {
			return nil, nil
		}
		var err error
		simplifiedType, err = NewSimplifiedTypeFromType(typeDef)
		if err != nil {
			return nil, err
		}
		m.SimplifiedTypes[name] = simplifiedType
	}
	return simplifiedType, nil
}

// Returns the full type definition for the type with the specified name
func (m *Schema) GetType(name string) *ast.Definition {
	if typeDef, ok := m.Schema.Types[name]; ok {
		return typeDef
	}
	return nil
}

// Adds/Updates the provided interface to the schema
func (m *Schema) SetInterface(simplifiedInterface *SimplifiedInterface) {
	m.Schema.Types[simplifiedInterface.Name] = CreateInterface(simplifiedInterface)
}

// Updates the schema with the provided type, if the provided type already
// exists it campares it with the current one to determine the differences
// and update the schema accordingly
// Shouldn't need tp worry about interfaces as they have already been applied
// should be same as old in case its not a new type
func (m *Schema) UpdateType(newType *SimplifiedType) (SchemaUpdateOp, error) {
	oldType, err := m.GetSimplifiedType(newType.Name)
	if err != nil {
		return SchemaUpdateOp_None, err
	}
	// fmt.Println("OldType: ", oldType)
	if oldType == nil {
		m.Schema.Types[newType.Name] = CreateType(newType)
		m.SimplifiedTypes[newType.Name] = newType.Clone()
		return SchemaUpdateOp_Created, nil
	}
	toAdd, toUpdate, err := oldType.PrepareFieldUpdate(newType)
	if err != nil {
		return SchemaUpdateOp_None, err
	}
	updateOp := SchemaUpdateOp_None
	if len(toAdd) > 0 || len(toUpdate) > 0 {
		fieldDefs := &m.GetType(newType.Name).Fields
		for _, field := range toUpdate {
			pos := findFieldPos(field.Name, *fieldDefs)
			(*fieldDefs)[pos] = CreateField(field)
			oldType.Fields[field.Name] = field
		}
		for _, field := range toAdd {
			*fieldDefs = append(*fieldDefs, CreateField(field))
			oldType.Fields[field.Name] = field
		}
		updateOp = SchemaUpdateOp_Updated
	}
	// fmt.Println("toAdd: ", toAdd)
	// fmt.Println("toUpdate: ", toUpdate)
	// fmt.Println("interfaces: ", newInterfaces)
	return updateOp, nil
}

// Adds/Updates the schema with the provided edge
func (m *Schema) UpdateEdge(typeName, edgeName, edgeType string) (bool, error) {
	return m.UpdateField(typeName, NewEdgeField(edgeName, edgeType))
}

// Adds/Updates the schema with the specified field
func (m *Schema) UpdateField(typeName string, field *SimplifiedField) (bool, error) {
	if field.NonNull {
		return false, fmt.Errorf("can't add non null field: %v to type: %v", field.Name, typeName)
	}
	typeDef := m.GetType(typeName)
	if typeDef == nil {
		return false, fmt.Errorf("failed to update field, definition for type: %v not found", typeName)
	}
	simplifiedType, err := m.GetSimplifiedType(typeName)
	if err != nil {
		return false, fmt.Errorf("failed to update field, simplified type: %v not found", typeName)
	}
	fieldDefs := &typeDef.Fields
	currentField := simplifiedType.Fields[field.Name]
	if currentField == nil {
		*fieldDefs = append(*fieldDefs, CreateField(field))
		simplifiedType.Fields[field.Name] = field
		return true, nil
	} else {
		if !currentField.equal(field) {
			err = currentField.CheckUpdate(field)
			if err != nil {
				return false, fmt.Errorf("can't update type: %v, error: %v", typeName, err)
			}
			pos := findFieldPos(field.Name, *fieldDefs)
			(*fieldDefs)[pos] = CreateField(field)
			simplifiedType.Fields[field.Name] = field
			return true, nil
		}
	}
	return false, nil
}

// Returns whether the type extends from document
func ExtendsDocument(typeDef *ast.Definition) bool {
	return HasInterface(typeDef, "Document")
}

// Returns whether the type extends the specified interface
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

func (m *Schema) String() string {
	out := &strings.Builder{}
	fmttr := formatter.NewFormatter(out)
	fmttr.FormatSchema(m.Schema)
	return out.String()
}

// Creates a fully defined type from the provided simplified type
func CreateType(simplifiedType *SimplifiedType) *ast.Definition {
	interfaces := []string{}
	interfaces = append(interfaces, simplifiedType.Interfaces...)

	def := CreateBaseType(ast.Object, simplifiedType.SimplifiedBaseType)
	def.Interfaces = interfaces
	return def
}

// Creates a fully defined interface from the provided simplified interface
func CreateInterface(simplifiedInterface *SimplifiedInterface) *ast.Definition {
	return CreateBaseType(ast.Interface, simplifiedInterface.SimplifiedBaseType)
}

// Creates a fully defined base type from the provided simplified base type
func CreateBaseType(kind ast.DefinitionKind, simplifiedBaseType *SimplifiedBaseType) *ast.Definition {
	var fieldDefs ast.FieldList
	fieldDefs = addFields(simplifiedBaseType.Fields, fieldDefs)
	var directives ast.DirectiveList
	if simplifiedBaseType.WithSubscription {
		directives = append(directives, &ast.Directive{
			Name: "withSubscription",
		})
	}

	return &ast.Definition{
		Kind:       kind,
		Name:       simplifiedBaseType.Name,
		Fields:     fieldDefs,
		Directives: directives,
	}
}

func addFields(fields map[string]*SimplifiedField, fieldList ast.FieldList) ast.FieldList {
	for _, field := range fields {
		fieldList = append(fieldList, CreateField(field))
	}
	return fieldList
}

// Creates a fully defined field from the provided simplified field
func CreateField(field *SimplifiedField) *ast.FieldDefinition {

	fieldType := &ast.Type{
		NonNull: field.NonNull,
	}
	if field.IsArray {
		fieldType.Elem = &ast.Type{
			NamedType: field.Type,
			NonNull:   true,
		}
	} else {
		fieldType.NamedType = field.Type
	}
	var directives ast.DirectiveList
	if field.IsID {
		directives = append(directives, &ast.Directive{
			Name: "id",
		})
	}
	if field.Indexes.HasIndexes() {
		children := make(ast.ChildValueList, 0, field.Indexes.Len())
		for index := range field.Indexes {
			children = append(
				children,
				&ast.ChildValue{
					Value: &ast.Value{
						Raw:  index,
						Kind: ast.EnumValue,
					},
				},
			)
		}
		directives = append(directives, &ast.Directive{
			Name: "search",
			Arguments: ast.ArgumentList{
				{
					Name: "by",
					Value: &ast.Value{
						Kind:     ast.ListValue,
						Children: children,
					},
				},
			},
		})

	}
	return &ast.FieldDefinition{
		Name:       field.Name,
		Type:       fieldType,
		Directives: directives,
	}
}

func DefinitionToString(def *ast.Definition, depth int) string {
	out := &strings.Builder{}
	out.WriteString(fmt.Sprintf("%vDefinition Name:%v\n", indent(depth), def.Name))
	out.WriteString(fmt.Sprintf("%vInterfaces: %v\n", indent(depth+1), def.Interfaces))
	out.WriteString(fmt.Sprintf("%vTypes: %v\n\n", indent(depth+1), def.Types))
	out.WriteString(fmt.Sprintf("%vDirectives: \n", indent(depth+1)))
	out.WriteString(DirectivesToString(def.Directives, depth+2))
	out.WriteString(fmt.Sprintf("%vFields: \n", indent(depth+1)))
	out.WriteString(FieldsToString(def.Fields, depth+2))
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
