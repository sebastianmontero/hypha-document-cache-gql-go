package gql

import (
	"fmt"
)

// Stores the data associated to an instance and provides the
// functionality to manage it
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

// Returns the values for the updatable fields, basically the values for non id fields
func (m *SimplifiedInstance) GetUpdateValues() (map[string]interface{}, error) {
	values := make(map[string]interface{}, len(m.Values))
	for name, value := range m.Values {
		if !m.SimplifiedType.GetField(name).IsID {
			values[name] = value
		}
	}
	return values, nil
}

// Returns the values that have to be removed, no longer exist on the new instance
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

// Returns the mutation to add this instance to the db
func (m *SimplifiedInstance) AddMutation(upsert bool) *Mutation {
	return m.SimplifiedType.AddMutation(m.Values, upsert)
}

// Returns the mutation to update this instance in the db
func (m *SimplifiedInstance) UpdateMutation(idName string, oldInstance *SimplifiedInstance) (*Mutation, error) {
	idValue, err := m.GetIdValue(idName)
	if err != nil {
		return nil, fmt.Errorf("failed creating update mutation, err: %v", err)
	}

	set, err := m.GetUpdateValues()
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

// Returns the mutation to delete this instance from the db
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
