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

func (m *Client) GetOne(id interface{}, simplifiedType *SimplifiedType, projection []string) (*SimplifiedInstance, error) {

	instances, err := m.Get([]interface{}{id}, simplifiedType, projection)
	if err != nil {
		return nil, err
	}
	return instances[id], nil
}

func (m *Client) Get(ids []interface{}, simplifiedType *SimplifiedType, projection []string) (map[interface{}]*SimplifiedInstance, error) {

	id, err := simplifiedType.GetIdField()
	if err != nil {
		return nil, fmt.Errorf("failed getting: %v with ids: %v error: %v", simplifiedType.Name, ids, err)
	}
	//Make sure id is in the projection
	if projection != nil {
		projection = append(projection, id.Name)
	}
	queryName, query, err := simplifiedType.GetStmt(projection)
	if err != nil {
		return nil, fmt.Errorf("failed getting: %v with ids: %v error: %v", simplifiedType.Name, ids, err)
	}
	req := graphql.NewRequest(query)
	req.Var("ids", ids)
	var response interface{}
	err = m.client.Run(context.Background(), req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed getting: %v with ids: %v, query:%v, error: %v", simplifiedType.Name, ids, query, err)
	}
	data := response.(map[string]interface{})[queryName].([]interface{})
	// fmt.Println("Response: ", response)
	instances := make(map[interface{}]*SimplifiedInstance, len(ids))
	for _, values := range data {
		v := values.(map[string]interface{})
		instances[v[id.Name]] = &SimplifiedInstance{
			SimplifiedType: simplifiedType,
			Values:         v,
		}
	}
	return instances, nil
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

func (m *Client) Add(simplifiedInstance *SimplifiedInstance, upsert bool) error {
	simplifiedType := simplifiedInstance.SimplifiedType

	query := simplifiedType.AddStmt()
	req := graphql.NewRequest(query)
	req.Var("input", simplifiedInstance.Values)
	req.Var("upsert", upsert)

	err := m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed inserting: %v, values: %v, stmt: %v error: %v", simplifiedType.Name, simplifiedInstance.Values, query, err)
	}
	return nil
}

// func (m *Client) UpdateSet(newInstance *SimplifiedInstance) error {
// 	return m.Update(newInstance, nil)
// }

func (m *Client) UpdateInstance(newInstance *SimplifiedInstance, oldInstance *SimplifiedInstance) error {
	idValue, err := newInstance.GetIdValue()
	if err != nil {
		return fmt.Errorf("failed updating, err: %v", err)
	}
	simplifiedType := newInstance.SimplifiedType
	query, err := simplifiedType.UpdateStmt()
	if err != nil {
		return err
	}
	req := graphql.NewRequest(query)
	req.Var("id", idValue)
	set, err := newInstance.GetUpdateValues()
	if err != nil {
		return err
	}
	req.Var("set", set)
	var remove map[string]interface{}
	if oldInstance != nil {
		remove = oldInstance.GetRemoveValues(newInstance)
	} else {
		remove = make(map[string]interface{})
	}
	req.Var("remove", remove)
	// fmt.Println("remove: ", remove)
	err = m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed updating: %v, set values: %v, remove values: %v, stmt: %v error: %v", simplifiedType.Name, set, remove, query, err)
	}
	return nil
}

func (m *Client) Update(id interface{}, set, remove map[string]interface{}, simplifiedType *SimplifiedType) error {

	query, err := simplifiedType.UpdateStmt()
	if err != nil {
		return err
	}
	req := graphql.NewRequest(query)
	req.Var("id", id)
	req.Var("set", set)
	req.Var("remove", remove)
	err = m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed updating: %v, set values: %v, remove values: %v, stmt: %v error: %v", simplifiedType.Name, set, remove, query, err)
	}
	return nil
}

func (m *Client) Delete(simplifiedInstance *SimplifiedInstance) error {

	idValue, err := simplifiedInstance.GetIdValue()
	if err != nil {
		return fmt.Errorf("failed deleting, err: %v", err)
	}
	simplifiedType := simplifiedInstance.SimplifiedType
	query, err := simplifiedType.DeleteStmt()
	if err != nil {
		return err
	}
	req := graphql.NewRequest(query)
	req.Var("id", idValue)
	err = m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("failed deleting: %v, id: %v, stmt: %v error: %v", simplifiedType.Name, idValue, query, err)
	}
	return nil
}
