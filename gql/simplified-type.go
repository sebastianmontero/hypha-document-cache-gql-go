package gql

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

type toFieldStmt func(field *SimplifiedField) string

type SimplifiedType struct {
	Name            string
	Fields          map[string]*SimplifiedField
	ExtendsDocument bool
}

func NewSimplifiedType(typeDef *ast.Definition) (*SimplifiedType, error) {
	fields := make(map[string]*SimplifiedField)

	for _, fieldDef := range typeDef.Fields {
		field, err := NewSimplifiedField(fieldDef)
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

func (m *SimplifiedType) Clone() *SimplifiedType {
	fields := make(map[string]*SimplifiedField, len(m.Fields))
	for name, field := range m.Fields {
		fields[name] = field
	}
	return &SimplifiedType{
		Name:            m.Name,
		Fields:          fields,
		ExtendsDocument: m.ExtendsDocument,
	}
}

func (m *SimplifiedType) GetIdField() (*SimplifiedField, error) {
	if m.ExtendsDocument {
		return DocumentFieldArgs["hash"], nil
	} else {
		for _, field := range m.Fields {
			if field.IsID {
				return field, nil
			}
		}
	}
	return nil, fmt.Errorf("type: %v has no id field", m.Name)
}

func (m *SimplifiedType) GetField(name string) *SimplifiedField {
	if field, ok := m.Fields[name]; ok {
		return field
	}
	return nil
}

func (m *SimplifiedType) PrepareUpdate(new *SimplifiedType) (toAdd []*SimplifiedField, toUpdate []*SimplifiedField, err error) {
	if new.ExtendsDocument && !m.ExtendsDocument {
		err = fmt.Errorf("can't add Document interface to type: %v", m.Name)
		return
	}
	if !new.ExtendsDocument && m.ExtendsDocument {
		err = fmt.Errorf("can't remove Document interface to type: %v", m.Name)
		return
	}
	toAdd = make([]*SimplifiedField, 0)
	toUpdate = make([]*SimplifiedField, 0)
	for _, field := range new.Fields {
		oldField := m.GetField(field.Name)
		if oldField == nil {
			if field.NonNull {
				err = fmt.Errorf("can't add non null field: %v to type: %v", field.Name, m.Name)
				return
			}
			toAdd = append(toAdd, field)
		} else {
			if *oldField != *field {
				err = oldField.CheckUpdate(field)
				if err != nil {
					err = fmt.Errorf("can't update type: %v, error: %v", m.Name, err)
				}
				toUpdate = append(toUpdate, field)
			}
		}
	}
	return
}

func (m *SimplifiedType) String() string {
	return fmt.Sprintf(
		`
			SimplifiedType: {
				Name: %v,
				Fields: %v,
				ExtendsDocument: %v,
			}		
		`,
		m.Name,
		m.Fields,
		m.ExtendsDocument,
	)
}

func (m *SimplifiedType) GetStmt(projection []string) (string, string, error) {

	id, err := m.GetIdField()
	if err != nil {
		return "", "", err
	}
	queryName := fmt.Sprintf("query%v", m.Name)
	docFields := ""
	if m.ExtendsDocument {
		docFields = queryFieldsStmt(DocumentFieldArgs, projection)
	}
	stmt := fmt.Sprintf(
		`
			query($ids: [%v!]!){
				%v(filter: { %v }){
					%v
					%v
				}
			}
		`,
		id.Type,
		queryName,
		inFilterStmt(id.Name, "ids"),
		docFields,
		queryFieldsStmt(m.Fields, projection),
	)

	return queryName, stmt, nil
}

// func toMap(values []string) map[string]bool {
// 	if values == nil {
// 		return nil
// 	}
// 	m := make(map[string]bool, len(values))
// 	for _, value := range values {
// 		m[value] = true
// 	}
// 	return m
// }

// func (m *SimplifiedType) AddStmt() string {

// 	docParams := ""
// 	docInputs := ""
// 	if m.ExtendsDocument {
// 		docParams = inputParamsStmt(DocumentFieldArgs, false)
// 		docInputs = inputFieldsStmt(DocumentFieldArgs, false)
// 	}
// 	return fmt.Sprintf(
// 		`
// 			mutation(
// 				%v
// 				%v
// 			) {
// 				add%v(input: [
// 					{
// 						%v
// 						%v
// 					}
// 				]){numUids}
// 			}
// 		`,
// 		docParams,
// 		inputParamsStmt(m.Fields, true),
// 		m.Name,
// 		docInputs,
// 		inputFieldsStmt(m.Fields, true),
// 	)
// }

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

func (m *SimplifiedType) UpdateMutation(id interface{}, set, remove map[string]interface{}) (*Mutation, error) {
	idField, err := m.GetIdField()
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
			idParamName:     id,
			setParamName:    set,
			removeParamName: remove,
		},
	}, nil
}

func (m *SimplifiedType) DeleteMutation(id interface{}) (*Mutation, error) {
	idField, err := m.GetIdField()
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
			idParamName: id,
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

// func inputFieldsStmt(fields map[string]*SimplifiedField, trimEnd bool) string {
// 	return join(fields, inputFieldStmt, ",\n", trimEnd)
// }

// func inputParamsStmt(fields map[string]*SimplifiedField, trimEnd bool) string {
// 	return join(fields, inputParamStmt, ",\n", trimEnd)
// }

func queryFieldStmt(field *SimplifiedField) string {
	if field.IsObject() {
		return fmt.Sprintf("%v{hash}", field.Name)
	} else {
		return fmt.Sprintf("%v", field.Name)
	}

}

// func inputFieldStmt(field *SimplifiedField) string {
// 	return fmt.Sprintf("%v: $%v", field.Name, field.Name)
// }

// func inputParamStmt(field *SimplifiedField) string {
// 	nonNull := ""
// 	if field.NonNull {
// 		nonNull = "!"
// 	}
// 	return fmt.Sprintf("$%v: %v%v", field.Name, field.Type, nonNull)
// }

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
