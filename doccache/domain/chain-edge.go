package domain

import (
	"fmt"
)

//ChainEdge domain object
type ChainEdge struct {
	Name string `json:"edge_name,omitempty"`
	From string `json:"from_node,omitempty"`
	To   string `json:"to_node,omitempty"`
}

// func (m *ChainEdge) UnmarshalJSON(b []byte) error {
// 	if err := json.Unmarshal(b, m); err != nil {
// 		return err
// 	}
// 	m.From = strings.ToUpper(m.From)
// 	m.To = strings.ToUpper(m.To)
// 	return nil
// }

func (m *ChainEdge) GetEdgeRef() map[string]interface{} {
	return map[string]interface{}{
		m.Name: []map[string]interface{}{{"hash": m.To}},
	}
}

func (m *ChainEdge) String() string {
	return fmt.Sprintf("ChainEdge{Name: %v, From: %v, To: %v}", m.Name, m.From, m.To)
}
