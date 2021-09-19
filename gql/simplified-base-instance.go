package gql

import (
	"fmt"
	"strconv"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/util"
)

type SimplifiedBaseInstance struct {
	SimplifiedBaseType *SimplifiedBaseType
	Values             map[string]interface{}
}

func NewSimplifiedBaseInstance(simplifiedBaseType *SimplifiedBaseType, values map[string]interface{}) *SimplifiedBaseInstance {
	return &SimplifiedBaseInstance{
		SimplifiedBaseType: simplifiedBaseType,
		Values:             values,
	}
}

func (m *SimplifiedBaseInstance) GetValue(name string) interface{} {
	if value, ok := m.Values[name]; ok {
		if value == nil {
			return nil
		}
		field := m.SimplifiedBaseType.GetField(name)
		switch field.Type {
		case GQLType_Int64:
			intValue, _ := strconv.ParseInt(fmt.Sprintf("%v", value), 10, 64)
			return intValue
		case GQLType_Time:
			return util.ToTime(fmt.Sprintf("%v", value))
		default:
			return value
		}
	}
	return nil
}

func (m *SimplifiedBaseInstance) SetValue(name string, value interface{}) {
	m.Values[name] = value
}

func (m *SimplifiedBaseInstance) GetIdValue(idName string) (interface{}, error) {
	id, err := m.SimplifiedBaseType.GetIdField(idName)
	if err != nil {
		return nil, fmt.Errorf("couldn't get id value, error: %v", err)
	}
	idValue, ok := m.Values[id.Name]
	if !ok {
		return nil, fmt.Errorf("no id value set for type: %v, values: %v", m.SimplifiedBaseType.Name, m.Values)
	}
	return idValue, nil
}

func (m *SimplifiedBaseInstance) String() string {
	return fmt.Sprintf(
		`
			SimplfiedBaseInstance {
				SimplifiedBaseType: %v
				Values: %v
			}
		`,
		m.SimplifiedBaseType,
		m.Values,
	)
}
