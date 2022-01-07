package domain

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
)

//ChainEdge domain object
type ChainEdge struct {
	Name        string `json:"edge_name,omitempty"`
	From        string `json:"from_node,omitempty"`
	To          string `json:"to_node,omitempty"`
	DocEdgeName string
}

func NewChainEdge(name, from, to string) *ChainEdge {
	return &ChainEdge{
		Name:        name,
		From:        from,
		To:          to,
		DocEdgeName: getDocEdgeName(name),
	}
}

func (m *ChainEdge) UnmarshalJSON(b []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(b, &data); err != nil {
		return err
	}
	m.Name = data["edge_name"].(string)
	m.From = strconv.FormatUint(uint64(data["from_node"].(float64)), 10)
	m.To = strconv.FormatUint(uint64(data["to_node"].(float64)), 10)
	m.DocEdgeName = getDocEdgeName(m.Name)
	return nil
}

func (m *ChainEdge) GetEdgeRef(docId interface{}) map[string]interface{} {
	return map[string]interface{}{
		m.DocEdgeName: []map[string]interface{}{{"docId": docId}},
	}
}

func (m *ChainEdge) String() string {
	return fmt.Sprintf("ChainEdge{Name: %v, From: %v, To: %v}", m.Name, m.From, m.To)
}

func getDocEdgeName(name string) string {
	return strcase.ToLowerCamel(strings.ReplaceAll(name, ".", "_"))
}
