package gql

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
)

type Admin struct {
	client *graphql.Client
}

func NewAdmin(endpoint string) *Admin {
	return &Admin{
		client: graphql.NewClient(endpoint),
	}
}

func (m *Admin) GetCurrentSchema() (*Schema, error) {
	req := graphql.NewRequest(`
		{
			getGQLSchema{
				schema
				generatedSchema
			}
		}
	`)
	var response interface{}
	err := m.client.Run(context.Background(), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed getting current schema, error: %v", err)
	}
	gqlSchema := response.(map[string]interface{})["getGQLSchema"]
	// fmt.Println("Response: ", response)
	if gqlSchema == nil {
		return nil, nil
	}
	schema := gqlSchema.(map[string]interface{})["schema"].(string)
	if schema == "" {
		return nil, nil
	}
	// fmt.Println("Response: ", gqlSchema.(map[string]interface{})["schema"].(string))
	return LoadSchema(schema)
}

func (m *Admin) UpdateSchema(schema *Schema) error {
	req := graphql.NewRequest(`
		mutation($schema: String!) {
			updateGQLSchema(
				input: {
					set: {
						schema:$schema
					}
				}
			){
				gqlSchema {id}
			}
		}
	`)
	req.Var("schema", schema.String())

	err := m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed updating schema, error: %v", err)
	}

	//Wait for schema to update
	_, err = m.GetCurrentSchema()
	if err != nil {
		return fmt.Errorf("failed updating schema, error getting updated schema: %v", err)
	}
	// fmt.Println("Health: ", health)

	return nil
}

func (m *Admin) Health() (string, error) {
	req := graphql.NewRequest(`
		{
			health{
				instance
				status
				ongoing
				indexing
			}
		}
	`)
	var response interface{}
	err := m.client.Run(context.Background(), req, &response)
	if err != nil {
		return "", fmt.Errorf("failed getting health state, error: %v", err)
	}
	return fmt.Sprintf("%v", response.(map[string]interface{})["health"]), nil
}
