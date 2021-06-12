package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
)

func TestChainEdgeUnmarshall(t *testing.T) {
	chainDocEdge := `{"contract":"dao.hypha","created_date":"2021-01-11T21:52:32","creator":"dao.hypha","edge_name":"settings","from_node":"52a7ff82bd6f53b31285e97d6806d886eefb650e79754784e9d923d3df347c91","from_node_edge_name_index":493623357,"from_node_to_node_index":340709097,"id":2475211255,"to_node":"3e06f9f93fb27ad04a2e97dfce9796c2d51b73721d6270e1c0ea6bf7e79c944b","to_node_edge_name_index":2119673673}`
	chainEdge := &domain.ChainEdge{}
	err := json.Unmarshal([]byte(chainDocEdge), chainEdge)
	if err != nil {
		t.Fatalf("Unmarshalling failed: %v", err)
	}
}
