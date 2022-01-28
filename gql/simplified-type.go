package gql

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

type toFieldStmt func(field *SimplifiedField) string

type SimplifiedType struct {
	*SimplifiedBaseType
	Interfaces []string
}

func NewSimplifiedType(name string, simplifiedFields map[string]*SimplifiedField, coreInterface *SimplifiedInterface) *SimplifiedType {

	simplifiedType := &SimplifiedType{
		SimplifiedBaseType: &SimplifiedBaseType{
			Name:             name,
			Fields:           make(map[string]*SimplifiedField),
			WithSubscription: true,
		},
		Interfaces: make([]string, 0),
	}
	simplifiedType.SetFields(simplifiedFields)
	if coreInterface != nil {
		simplifiedType.addCoreInterface(coreInterface)
	}
	return simplifiedType
}

func NewSimplifiedTypeFromType(typeDef *ast.Definition) (*SimplifiedType, error) {
	fields := make(map[string]*SimplifiedField)

	for _, fieldDef := range typeDef.Fields {
		field, err := NewSimplifiedField(fieldDef)
		if err != nil {
			return nil, fmt.Errorf("failed to create simplified type from type definition for type: %v, error: %v", typeDef.Name, err)
		}
		fields[field.Name] = field
	}
	interfaces := make([]string, len(typeDef.Interfaces))
	copy(interfaces, typeDef.Interfaces)

	return &SimplifiedType{
		SimplifiedBaseType: &SimplifiedBaseType{
			Name:             typeDef.Name,
			Fields:           fields,
			WithSubscription: typeDef.Directives.ForName("withSubscription") != nil,
		},
		Interfaces: interfaces,
	}, nil
}

func (m *SimplifiedType) SetFields(fields map[string]*SimplifiedField) {
	for name, field := range fields {
		m.Fields[name] = field
	}
}

func (m *SimplifiedType) SetFieldArray(fields []*SimplifiedField) {
	for _, field := range fields {
		m.Fields[field.Name] = field
	}
}

func (m *SimplifiedType) addCoreInterface(coreInterface *SimplifiedInterface) {
	m.Interfaces = append(m.Interfaces, coreInterface.Name)
	m.SetFields(coreInterface.Fields)
}

func (m *SimplifiedType) AddInterface(simplifiedInterface *SimplifiedInterface) error {
	toAdd, toUpdate, err := m.PrepareInterfaceFieldUpdate(simplifiedInterface)
	if err != nil {
		return fmt.Errorf("failed to apply interface, error: %v", err)
	}
	m.Interfaces = append(m.Interfaces, simplifiedInterface.Name)
	m.SetFieldArray(toAdd)
	m.SetFieldArray(toUpdate)
	return nil
}

func (m *SimplifiedType) HasInterface(name string) bool {
	for _, interf := range m.Interfaces {
		if interf == name {
			return true
		}
	}
	return false
}

// func (m *SimplifiedType) AddInterfaces(simplifiedInterfaces SimplifiedInterfaces) {
// 	for _, simplifiedInterface := range simplifiedInterfaces {
// 		m.AddInterface(simplifiedInterface)
// 	}
// }

func (m *SimplifiedType) Clone() *SimplifiedType {
	fields := make(map[string]*SimplifiedField, len(m.Fields))
	for name, field := range m.Fields {
		fields[name] = field
	}
	return &SimplifiedType{
		SimplifiedBaseType: &SimplifiedBaseType{
			Name:             m.Name,
			Fields:           fields,
			WithSubscription: m.WithSubscription,
		},
		Interfaces: m.CloneInterfaces(),
	}
}

func (m *SimplifiedType) CloneInterfaces() []string {
	interfaces := make([]string, len(m.Interfaces))
	copy(interfaces, m.Interfaces)
	return interfaces
}

func (m *SimplifiedType) PrepareFieldUpdate(new *SimplifiedType) (toAdd []*SimplifiedField, toUpdate []*SimplifiedField, err error) {
	return m.SimplifiedBaseType.PrepareFieldUpdate(new.SimplifiedBaseType)
}

func (m *SimplifiedType) PrepareInterfaceFieldUpdate(simplifiedInterface *SimplifiedInterface) (toAdd []*SimplifiedField, toUpdate []*SimplifiedField, err error) {
	return m.SimplifiedBaseType.PrepareFieldUpdate(simplifiedInterface.SimplifiedBaseType)
}

func (m *SimplifiedType) PrepareInterfaceUpdate(new *SimplifiedType) []string {
	toAdd := make([]string, 0)
	for _, interf := range new.Interfaces {
		if !m.HasInterface(interf) {
			toAdd = append(toAdd, interf)
		}
	}
	return toAdd
}

func (m *SimplifiedType) String() string {
	return fmt.Sprintf(
		`
			SimplifiedType: {
				Name: %v,
				Fields: %v,
				Interfaces: %v,
			}		
		`,
		m.Name,
		m.Fields,
		m.Interfaces,
	)
}

func (m *SimplifiedType) AddMutation(values map[string]interface{}, upsert bool) *Mutation {
	inputParamName := m.addInputParamNameStmt()
	upsertParamName := m.upsertParamNameStmt()
	return &Mutation{
		ParamStmt: fmt.Sprintf(
			"$%v: %v, $%v: Boolean",
			inputParamName,
			m.addInputParamTypeStmt(),
			upsertParamName,
		),
		MutationStmt: fmt.Sprintf(
			"add%v(input: $%v, upsert: $%v){numUids}",
			m.Name,
			inputParamName,
			upsertParamName,
		),
		Params: map[string]interface{}{
			inputParamName:  values,
			upsertParamName: upsert,
		},
	}
}

func (m *SimplifiedType) addInputParamTypeStmt() string {
	return fmt.Sprintf(
		"[Add%vInput!]!",
		m.Name,
	)
}

func (m *SimplifiedType) addInputParamNameStmt() string {
	return m.nameStmt("input")
}

func (m *SimplifiedType) upsertParamNameStmt() string {
	return m.nameStmt("upsert")
}

func (m *SimplifiedType) nameStmt(param string) string {
	return fmt.Sprintf(
		"%v%v",
		param,
		m.Name,
	)
}

func (m *SimplifiedType) UpdateMutation(idName string, idValue interface{}, set, remove map[string]interface{}) (*Mutation, error) {
	idField, err := m.GetIdField(idName)
	if err != nil {
		return nil, err
	}
	idParamName := m.idParamNameStmt()
	setParamName := m.setParamNameStmt()
	removeParamName := m.removeParamNameStmt()
	patchParamType := m.patchParamTypeStmt()
	return &Mutation{
		ParamStmt: fmt.Sprintf(
			"$%v: %v!, $%v: %v, $%v: %v",
			idParamName,
			idField.Type,
			setParamName,
			patchParamType,
			removeParamName,
			patchParamType,
		),
		MutationStmt: fmt.Sprintf(
			"update%v(input: { filter: { %v }, set: $%v, remove: $%v }){numUids}",
			m.Name,
			eqFilterStmt(idField.Name, idParamName),
			setParamName,
			removeParamName,
		),
		Params: map[string]interface{}{
			idParamName:     idValue,
			setParamName:    set,
			removeParamName: remove,
		},
	}, nil
}

func (m *SimplifiedType) DeleteMutation(idName string, idValue interface{}) (*Mutation, error) {
	idField, err := m.GetIdField(idName)
	if err != nil {
		return nil, err
	}
	idParamName := m.idParamNameStmt()
	return &Mutation{
		ParamStmt: fmt.Sprintf(
			"$%v: %v!",
			idParamName,
			idField.Type,
		),
		MutationStmt: fmt.Sprintf(
			"delete%v(filter: { %v }){numUids}",
			m.Name,
			eqFilterStmt(idField.Name, idParamName),
		),
		Params: map[string]interface{}{
			idParamName: idValue,
		},
	}, nil
}

func (m *SimplifiedType) idParamNameStmt() string {
	return m.nameStmt("id")
}

func (m *SimplifiedType) setParamNameStmt() string {
	return m.nameStmt("set")
}

func (m *SimplifiedType) removeParamNameStmt() string {
	return m.nameStmt("remove")
}

func (m *SimplifiedType) patchParamTypeStmt() string {
	return fmt.Sprintf(
		"%vPatch",
		m.Name,
	)
}

func eqFilterStmt(field, param string) string {
	return filterStmt(field, "eq", param)
}

func inFilterStmt(field, param string) string {
	return filterStmt(field, "in", param)
}

func filterStmt(field, op, param string) string {
	return fmt.Sprintf("%v: { %v: $%v }", field, op, param)
}

func queryFieldsStmt(fields map[string]*SimplifiedField, projection []string) string {
	var filtered map[string]*SimplifiedField
	if projection == nil {
		filtered = fields
	} else {
		filtered = make(map[string]*SimplifiedField)
		for _, fieldName := range projection {
			if field, ok := fields[fieldName]; ok {
				filtered[fieldName] = field
			}
		}
	}
	return join(filtered, queryFieldStmt, "\n", false)
}

func queryFieldStmt(field *SimplifiedField) string {
	if field.IsObject() {
		return fmt.Sprintf("%v{docId}", field.Name)
	} else {
		return fmt.Sprintf("%v", field.Name)
	}

}

func join(fields map[string]*SimplifiedField, fn toFieldStmt, separator string, trimEnd bool) string {
	q := &strings.Builder{}
	for _, field := range fields {
		q.WriteString(fmt.Sprintf("%v%v", fn(field), separator))
	}
	stmt := q.String()
	if trimEnd {
		stmt = strings.TrimSuffix(q.String(), separator)
	}
	return stmt
}
