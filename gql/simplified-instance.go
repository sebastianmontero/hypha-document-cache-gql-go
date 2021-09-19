package gql

import (
	"fmt"
)

type SimplifiedInstance struct {
	*SimplifiedBaseInstance
	SimplifiedType *SimplifiedType
}

func NewSimplifiedInstance(simplifiedType *SimplifiedType, values map[string]interface{}) *SimplifiedInstance {
	return &SimplifiedInstance{
		SimplifiedBaseInstance: NewSimplifiedBaseInstance(simplifiedType.SimplifiedBaseType, values),
		SimplifiedType:         simplifiedType,
	}
}

//idName: main non mutable id field
func (m *SimplifiedInstance) GetUpdateValues(idName string) (map[string]interface{}, error) {
	id, err := m.SimplifiedType.GetIdField(idName)
	if err != nil {
		return nil, fmt.Errorf("couldn't get update values, error: %v", err)
	}
	values := make(map[string]interface{}, len(m.Values))
	for name, value := range m.Values {
		if name != id.Name {
			values[name] = value
		}
	}
	return values, nil
}

func (m *SimplifiedInstance) GetRemoveValues(oldInstance *SimplifiedInstance) map[string]interface{} {
	remove := make(map[string]interface{})
	for name, value := range oldInstance.Values {
		if _, ok := m.Values[name]; !ok {
			if !oldInstance.SimplifiedType.Fields[name].IsEdge() {
				remove[name] = value
			}
		}
	}
	return remove
}

func (m *SimplifiedInstance) AddMutation(upsert bool) *Mutation {
	return m.SimplifiedType.AddMutation(m.Values, upsert)
}

func (m *SimplifiedInstance) UpdateMutation(idName string, oldInstance *SimplifiedInstance) (*Mutation, error) {
	idValue, err := m.GetIdValue(idName)
	if err != nil {
		return nil, fmt.Errorf("failed creating update mutation, err: %v", err)
	}

	set, err := m.GetUpdateValues(idName)
	if err != nil {
		return nil, err
	}
	var remove map[string]interface{}
	if oldInstance != nil {
		remove = m.GetRemoveValues(oldInstance)
	} else {
		remove = make(map[string]interface{})
	}

	return m.SimplifiedType.UpdateMutation(idName, idValue, set, remove)
}

func (m *SimplifiedInstance) DeleteMutation(idName string) (*Mutation, error) {
	idValue, err := m.GetIdValue(idName)
	if err != nil {
		return nil, fmt.Errorf("failed creating update mutation, err: %v", err)
	}
	return m.SimplifiedType.DeleteMutation(idName, idValue)
}

func (m *SimplifiedInstance) String() string {
	return fmt.Sprintf(
		`
			SimplfiedInstance {
				SimplifiedBaseInstance: %v
				SimplifiedType: %v
			}
		`,
		m.SimplifiedBaseInstance,
		m.SimplifiedType,
	)
}
