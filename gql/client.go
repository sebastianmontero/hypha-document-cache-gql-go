package gql

import (
	"context"
	"fmt"

	"github.com/machinebox/graphql"
)

type Client struct {
	client *graphql.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		client: graphql.NewClient(endpoint),
	}
}

func (m *Client) Get(hash string, simplifiedType *SimplifiedType, projection []string) (*SimplifiedInstance, error) {

	queryName, query := simplifiedType.GetStmt(hash, projection)
	req := graphql.NewRequest(query)
	req.Var("hash", hash)
	var response interface{}
	err := m.client.Run(context.Background(), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed getting: %v with hash: %v, query:%v, error: %v", simplifiedType.Name, hash, query, err)
	}
	values := response.(map[string]interface{})[queryName]
	fmt.Println("Response: ", response)
	if values == nil {
		return nil, nil
	}
	return &SimplifiedInstance{
		SimplifiedType: simplifiedType,
		Values:         values.(map[string]interface{}),
	}, nil
}

// func (m *Client) Add(simplifiedInstance *SimplifiedInstance) error {
// 	simplifiedType := simplifiedInstance.SimplifiedType

// 	query := simplifiedType.AddStmt()
// 	req := graphql.NewRequest(query)
// 	for name, value := range simplifiedInstance.Values {
// 		req.Var(name, value)
// 	}

// 	err := m.client.Run(context.Background(), req, nil)
// 	if err != nil {
// 		return fmt.Errorf("failed inserting: %v, values: %v, stmt: %v error: %v", simplifiedType.Name, simplifiedInstance.Values, query, err)
// 	}
// 	return nil
// }

func (m *Client) Add(simplifiedInstance *SimplifiedInstance) error {
	simplifiedType := simplifiedInstance.SimplifiedType

	query := simplifiedType.AddStmt()
	req := graphql.NewRequest(query)
	req.Var("input", simplifiedInstance.Values)

	err := m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed inserting: %v, values: %v, stmt: %v error: %v", simplifiedType.Name, simplifiedInstance.Values, query, err)
	}
	return nil
}
