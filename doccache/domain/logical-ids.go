package domain

import (
	"fmt"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
)

type LogicalIds map[string][]string

func NewLogicalIds() LogicalIds {
	return make(LogicalIds)
}

func (m LogicalIds) Set(typeName string, ids []string) {
	m[typeName] = ids
}

func (m LogicalIds) ConfigureLogicalIds(simplifiedBaseType *gql.SimplifiedBaseType) error {
	if ids, ok := m[simplifiedBaseType.Name]; ok {
		for _, id := range ids {
			field := simplifiedBaseType.GetField(id)
			if field != nil {
				field.IsID = true
				field.NonNull = true
			} else {
				return fmt.Errorf("failed configuring logical ids, type: %v does not have logical id field: %v", simplifiedBaseType.Name, id)
			}
		}
	}
	return nil
}
