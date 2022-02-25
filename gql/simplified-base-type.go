package gql

import "fmt"

type SimplifiedBaseType struct {
	Name             string
	WithSubscription bool
	Fields           map[string]*SimplifiedField
}

func NewSimplifiedBaseType(name string, fields map[string]*SimplifiedField) *SimplifiedBaseType {
	return &SimplifiedBaseType{
		Name:             name,
		WithSubscription: true,
		Fields:           fields,
	}
}

func (m *SimplifiedBaseType) GetObjectFields() []*SimplifiedField {
	objFields := make([]*SimplifiedField, 0)
	for _, field := range m.Fields {
		if field.IsObject() {
			objFields = append(objFields, field)
		}
	}
	return objFields
}

func (m *SimplifiedBaseType) GetIdField(name string) (*SimplifiedField, error) {
	idField := m.GetField(name)
	if idField == nil {
		return nil, fmt.Errorf("type: %v does not have field: %v", m.Name, name)
	}
	if idField.IsID {
		return idField, nil
	} else {
		return nil, fmt.Errorf("field: %v in type: %v is not an ID", name, m.Name)
	}
}

func (m *SimplifiedBaseType) GetCoreFields() []string {
	coreFields := make([]string, 0)
	for name, field := range m.Fields {
		if !field.IsEdge() {
			coreFields = append(coreFields, name)
		}
	}
	return coreFields
}

func (m *SimplifiedBaseType) GetField(name string) *SimplifiedField {
	if field, ok := m.Fields[name]; ok {
		return field
	}
	return nil
}

func (m *SimplifiedBaseType) HasField(name string) bool {
	_, ok := m.Fields[name]
	return ok
}

func (m *SimplifiedBaseType) SetField(name string, field *SimplifiedField) {
	m.Fields[name] = field
}

func (m *SimplifiedBaseType) GetStmt(idName string, projection []string) (string, string, error) {

	id, err := m.GetIdField(idName)
	if err != nil {
		return "", "", err
	}
	queryName := fmt.Sprintf("query%v", m.Name)
	stmt := fmt.Sprintf(
		`
			query($ids: [%v!]!){
				%v(filter: { %v }){
					%v
				}
			}
		`,
		id.Type,
		queryName,
		inFilterStmt(id.Name, "ids"),
		queryFieldsStmt(m.Fields, projection),
	)

	return queryName, stmt, nil
}

func (m *SimplifiedBaseType) PrepareFieldUpdate(new *SimplifiedBaseType) (toAdd []*SimplifiedField, toUpdate []*SimplifiedField, err error) {
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
			if !oldField.equal(field) {
				err = oldField.CheckUpdate(field)
				if err != nil {
					err = fmt.Errorf("can't update type: %v, error: %v", m.Name, err)
					return
				}
				toUpdate = append(toUpdate, field)
			}
		}
	}
	return
}
