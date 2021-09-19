package gql_test

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
)

// func TestUpdateSchemaDelay(t *testing.T) {
// 	cl := graphql.NewClient("http://localhost:8080/admin")

// 	schemaDef := generateTypes(1, 2000)

// 	fmt.Println("Updating schema...")
// 	err := UpdateSchema(cl, schemaDef)
// 	assert.NilError(t, err)

// 	fmt.Println("Getting schema...")
// 	generatedSchema, err := GetCurrentSchema(cl)
// 	assert.NilError(t, err)
// 	fmt.Println("Schema: ", generatedSchema)
// 	assert.Assert(t, strings.Contains(generatedSchema, "updatePerson2000"))

// 	schemaDef += generateTypes(2001, 2000)
// 	assert.Assert(t, strings.Contains(schemaDef, "Person4000"))
// 	fmt.Println("Updating schema...")
// 	err = UpdateSchema(cl, schemaDef)
// 	assert.NilError(t, err)

// 	fmt.Println("Getting schema...")
// 	generatedSchema, err = GetCurrentSchema(cl)
// 	assert.NilError(t, err)

// 	assert.Assert(t, strings.Contains(generatedSchema, "updatePerson4000"))

// }

func generateTypes(start, count int) string {
	schemaDef := ""
	for i := start; i < start+count; i++ {
		schemaDef += fmt.Sprintf(`
			type Person%v {
				name: String @search(by: [term])
				createdAt: DateTime @search(by: [day])
				intValue: Int64 @search(by: [int64])
				picks: [Int64]
			}
		`, i)
	}
	return schemaDef
}

func GetCurrentSchema(cl *graphql.Client) (string, error) {
	req := graphql.NewRequest(`
		{
			getGQLSchema{
				schema
				generatedSchema
			}
		}
	`)
	var response interface{}
	err := cl.Run(context.Background(), req, &response)
	if err != nil {
		return "", fmt.Errorf("failed getting current schema, error: %v", err)
	}
	gqlSchema := response.(map[string]interface{})["getGQLSchema"]

	if gqlSchema == nil {
		return "", nil
	}
	schema := gqlSchema.(map[string]interface{})["generatedSchema"].(string)
	if schema == "" {
		return "", nil
	}

	return schema, nil
}

func UpdateSchema(cl *graphql.Client, schema string) error {
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
	req.Var("schema", schema)

	err := cl.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed updating schema, error: %v", err)
	}

	return nil
}
