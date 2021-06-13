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

func (m *SimplifiedType) GetStmt(hash string, projection []string) (string, string) {
	queryName := fmt.Sprintf("get%v", m.Name)
	docFields := ""
	if m.ExtendsDocument {
		docFields = queryFieldsStmt(DocumentFieldArgs, projection)
	}
	stmt := fmt.Sprintf(
		`
			query($hash: String!){
				%v(hash: $hash){
					%v
					%v
				}
			}
		`,
		queryName,
		docFields,
		queryFieldsStmt(m.Fields, projection),
	)

	return queryName, stmt
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

func (m *SimplifiedType) AddStmt() string {

	return fmt.Sprintf(
		`
			mutation($input: [Add%vInput!]!) {
				add%v(input: $input){numUids}
			}
		`,
		m.Name,
		m.Name,
	)
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
