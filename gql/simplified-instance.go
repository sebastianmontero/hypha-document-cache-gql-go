package gql

import (
	"fmt"
)

type SimplifiedInstance struct {
	SimplifiedType *SimplifiedType
	Values         map[string]interface{}
}

func (m *SimplifiedInstance) GetIdValue() (interface{}, error) {
	id, err := m.SimplifiedType.GetIdField()
	if err != nil {
		return nil, fmt.Errorf("couldn't get id value, error: %v", err)
	}
	idValue, ok := m.Values[id.Name]
	if !ok {
		return nil, fmt.Errorf("no id value set for type: %v, values: %v", m.SimplifiedType.Name, m.Values)
	}
	return idValue, nil
}

func (m *SimplifiedInstance) GetUpdateValues() (map[string]interface{}, error) {
	id, err := m.SimplifiedType.GetIdField()
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

func (m *SimplifiedInstance) GetRemoveValues(newInstance *SimplifiedInstance) map[string]interface{} {
	remove := make(map[string]interface{})
	for name, value := range m.Values {
		if _, ok := newInstance.Values[name]; !ok {
			if !m.SimplifiedType.Fields[name].IsEdge() {
				remove[name] = value
			}
		}
	}
	return remove
}

func (m *SimplifiedInstance) String() string {
	return fmt.Sprintf(
		`
			SimplfiedInstance {
				SimplifiedType: %v
				Values: %v
			}
		`,
		m.SimplifiedType,
		m.Values,
	)
}
