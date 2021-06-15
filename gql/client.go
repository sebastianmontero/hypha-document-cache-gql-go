package gql

import (
	"context"
	"fmt"
	"strings"

	"github.com/machinebox/graphql"
)

type Mutation struct {
	ParamStmt    string
	MutationStmt string
	Params       map[string]interface{}
}

func (m *Mutation) HasParams() bool {
	return len(m.Params) > 0
}

func (m *Mutation) String() string {
	return fmt.Sprintf(
		`
	    Mutation{
				ParamStmt: %v
				MutationStmt: %v,
				Params: %v
			}	
		`,
		m.ParamStmt,
		m.MutationStmt,
		m.Params,
	)
}

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

func (m *Client) Mutate(mutations ...*Mutation) error {
	paramStmt := &strings.Builder{}
	mutationsStmt := &strings.Builder{}
	for _, mutation := range mutations {
		if mutation.HasParams() {
			paramStmt.WriteString(mutation.ParamStmt)
			paramStmt.WriteString(",")
		}
		mutationsStmt.WriteString(mutation.MutationStmt)
		mutationsStmt.WriteString("\n")
	}
	stmt := fmt.Sprintf(
		`
			mutation(%v){
				%v
			}
		`,
		paramStmt.String(),
		mutationsStmt.String(),
	)
	req := graphql.NewRequest(stmt)
	for _, mutation := range mutations {
		for name, value := range mutation.Params {
			req.Var(name, value)
		}
	}
	err := m.client.Run(context.Background(), req, nil)
	if err != nil {
		return fmt.Errorf("mutation failed, stmt: %v, mutations: %v, error: %v", stmt, mutations, err)
	}
	return nil
}
