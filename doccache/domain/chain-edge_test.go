package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"gotest.tools/assert"
)

func TestChainEdgeUnmarshall(t *testing.T) {
	chainDocEdge := `{"contract":"dao.hypha","created_date":"2021-01-11T21:52:32","creator":"dao.hypha","edge_name":"settings","from_node":1,"from_node_edge_name_index":493623357,"from_node_to_node_index":340709097,"id":2475211255,"to_node":2,"to_node_edge_name_index":2119673673}`
	chainEdge := &domain.ChainEdge{}
	err := json.Unmarshal([]byte(chainDocEdge), chainEdge)
	if err != nil {
		t.Fatalf("Unmarshalling failed: %v", err)
	}
	assert.Equal(t, chainEdge.Name, "settings")
	assert.Equal(t, chainEdge.From, "1")
	assert.Equal(t, chainEdge.To, "2")
}
