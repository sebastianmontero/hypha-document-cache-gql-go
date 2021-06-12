package gql

import (
	"fmt"
)

type SimplifiedInstance struct {
	SimplifiedType *SimplifiedType
	Values         map[string]interface{}
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
