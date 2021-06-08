package doccache

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
)

var dg *dgraph.Dgraph
var doccache *Doccache

// TestMain will exec each test, one by one
func TestMain(m *testing.M) {
	beforeAll()
	// exec test and this returns an exit code to pass to os
	retCode := m.Run()
	afterAll()
	// If exit code is distinct of zero,
	// the test will be failed (red)
	os.Exit(retCode)
}

func beforeAll() {
	var err error
	dg, err = dgraph.New("")
	if err != nil {
		log.Fatal(err, "Unable to create dgraph")
	}
	err = dg.DropAll()
	if err != nil {
		log.Fatal(err, "Unable to drop all")
	}
	doccache, err = New(dg, nil)
	if err != nil {
		log.Fatal(err, "Failed creating docCache")
	}
}

func afterAll() {
	dg.Close()
}

func TestOpCycle(t *testing.T) {
	if doccache.Cursor.UID == "" {
		t.Fatalf("Cursor should have UID already, it should be initialized on the creation of the doccache")
	}

	createdDate := "2020-11-12T18:27:47.000"
	chainDoc1 := &ChainDocument{
		ID:          0,
		Hash:        "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
		CreatedDate: createdDate,
		Creator:     "dao.hypha",
		ContentGroups: [][]*ChainContent{
			{
				{
					Label: "root_node",
					Value: []interface{}{
						"name",
						"dao.hypha",
					},
				},
			},
		},
	}

	expectedDoc1 := &Document{
		Hash:        "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
		CreatedDate: ToTime(createdDate),
		Creator:     "dao.hypha",
		DType:       []string{"Document"},
		ContentGroups: []*ContentGroup{
			{
				ContentGroupSequence: 1,
				DType:                []string{"ContentGroup"},
				Contents: []*Content{
					{
						Label:           "root_node",
						Type:            "name",
						Value:           "dao.hypha",
						ContentSequence: 1,
						DType:           []string{"Content"},
					},
				},
			},
		},
	}

	t.Logf("Storing new document")
	cursor := "cursor1"
	err := doccache.StoreDocument(chainDoc1, cursor)
	if err != nil {
		t.Fatalf("StoreDocument failed: %v", err)
	}

	doc, err := doccache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	compareDocs(expectedDoc1, doc, t)
	validateCursor(cursor, t)
	t.Logf("Updating document certificates")

	certificationDate := "2020-11-12T20:27:47.000"
	chainDoc1.Certificates = []*ChainCertificate{
		{
			Certifier:         "sebastian",
			Notes:             "Sebastian's Notes",
			CertificationDate: certificationDate,
		},
	}
	expectedDoc1.Certificates = []*Certificate{
		{
			Certifier:             "sebastian",
			Notes:                 "Sebastian's Notes",
			CertificationDate:     ToTime(certificationDate),
			CertificationSequence: 1,
			DType:                 []string{"Certificate"},
		},
	}
	cursor = "cursor2"
	err = doccache.StoreDocument(chainDoc1, cursor)
	if err != nil {
		t.Fatalf("StoreDocument failed: %v", err)
	}

	doc, err = doccache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	compareDocs(expectedDoc1, doc, t)
	validateCursor(cursor, t)

	t.Logf("Updating document certificates 2")

	certificationDate = "2020-11-14T20:27:47.000"
	chainDoc1.Certificates = append(chainDoc1.Certificates, &ChainCertificate{
		Certifier:         "pedro",
		Notes:             "Pedro's Notes",
		CertificationDate: certificationDate,
	})
	expectedDoc1.Certificates = append(expectedDoc1.Certificates, &Certificate{
		Certifier:             "pedro",
		Notes:                 "Pedro's Notes",
		CertificationDate:     ToTime(certificationDate),
		CertificationSequence: 2,
		DType:                 []string{"Certificate"},
	})

	cursor = "cursor3"
	err = doccache.StoreDocument(chainDoc1, cursor)
	if err != nil {
		t.Fatalf("StoreDocument failed: %v", err)
	}

	doc, err = doccache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	compareDocs(expectedDoc1, doc, t)
	validateCursor(cursor, t)

	createdDate = "2020-11-12T22:09:12.000"
	startTime := "2021-04-01T15:50:54.291"
	chainDoc2 := &ChainDocument{
		ID:          1,
		Hash:        "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
		CreatedDate: createdDate,
		Creator:     "dao.hypha1",
		ContentGroups: [][]*ChainContent{
			{
				{
					Label: "member",
					Value: []interface{}{
						"name",
						"1onefiftyfor",
					},
				},
				{
					Label: "role",
					Value: []interface{}{
						"name",
						"dev",
					},
				},
				{
					Label: "start_time",
					Value: []interface{}{
						"time_point",
						startTime,
					},
				},
			},
			{
				{
					Label: "root",
					Value: []interface{}{
						"checksum256",
						"d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
					},
				},
				{
					Label: "vote_count",
					Value: []interface{}{
						"int64",
						89,
					},
				},
			},
		},
	}

	voteCount := int64(89)
	expectedDoc2 := &Document{
		Hash:        "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
		CreatedDate: ToTime(createdDate),
		Creator:     "dao.hypha1",
		DType:       []string{"Document"},
		ContentGroups: []*ContentGroup{
			{
				ContentGroupSequence: 1,
				DType:                []string{"ContentGroup"},
				Contents: []*Content{
					{
						Label:           "member",
						Type:            "name",
						Value:           "1onefiftyfor",
						ContentSequence: 1,
						DType:           []string{"Content"},
					},
					{
						Label:           "role",
						Type:            "name",
						Value:           "dev",
						ContentSequence: 2,
						DType:           []string{"Content"},
					},
					{
						Label:           "start_time",
						Type:            "time_point",
						Value:           startTime,
						TimeValue:       ToTime(startTime),
						ContentSequence: 3,
						DType:           []string{"Content"},
					},
				},
			},
			{
				ContentGroupSequence: 2,
				DType:                []string{"ContentGroup"},
				Contents: []*Content{
					{
						Label:           "root",
						Type:            "checksum256",
						Value:           "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
						ContentSequence: 1,
						DType:           []string{"Content"},
						Document:        []*Document{expectedDoc1},
					},
					{
						Label:           "vote_count",
						Type:            "int64",
						Value:           "89",
						IntValue:        &voteCount,
						ContentSequence: 2,
						DType:           []string{"Content"},
					},
				},
			},
		},
	}

	cursor = "cursor4"
	t.Log("Storing another document")
	err = doccache.StoreDocument(chainDoc2, cursor)
	if err != nil {
		t.Fatalf("StoreDocument failed: %v", err)
	}

	doc, err = doccache.GetByHash("4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff", &RequestConfig{ContentGroups: true, Certificates: true})
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	compareDocs(expectedDoc2, doc, t)
	validateCursor(cursor, t)

	t.Log("Check original document wasn't modified")
	doc, err = doccache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	if err != nil {
		t.Fatalf("GetByHash failed: %v", err)
	}
	compareDocs(expectedDoc1, doc, t)

	t.Log("Adding edge")
	cursor = "cursor5"
	err = doccache.MutateEdge(&ChainEdge{
		Name: "member",
		From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
		To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	}, false, cursor)
	if err != nil {
		t.Fatalf("MutateEdge for adding failed: %v", err)
	}

	docAsMap, err := doccache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	if err != nil {
		t.Fatalf("GetByHashAsMap failed: %v", err)
	}
	t.Logf("Doc as map: %v", docAsMap)
	if docAsMap == nil {
		t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	}
	membersi, ok := docAsMap["member"]
	if !ok {
		t.Fatal("Expected to find member edge found none")
	}
	members := membersi.([]interface{})
	if len(members) != 1 {
		t.Fatalf("Expected to find 1 member found: %v", len(members))
	}
	member := members[0].(map[string]interface{})
	hashi, ok := member["hash"]
	if !ok {
		t.Fatal("Expected to find hash found none")
	}
	hash := hashi.(string)
	if hash != "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff" {
		t.Fatalf("Expected hash to be: 4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff found: %v", hash)
	}
	validateCursor(cursor, t)

	t.Log("Adding same edge")
	cursor = "cursor6"
	err = doccache.MutateEdge(&ChainEdge{
		Name: "member",
		From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
		To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	}, false, cursor)
	if err != nil {
		t.Fatalf("MutateEdge for adding failed: %v", err)
	}

	docAsMap, err = doccache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	if err != nil {
		t.Fatalf("GetByHashAsMap failed: %v", err)
	}
	t.Logf("Doc as map: %v", docAsMap)
	if docAsMap == nil {
		t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	}
	membersi, ok = docAsMap["member"]
	if !ok {
		t.Fatal("Expected to find member edge found none")
	}
	members = membersi.([]interface{})
	if len(members) != 1 {
		t.Fatalf("Expected to find 1 member found: %v", len(members))
	}
	member = members[0].(map[string]interface{})
	hashi, ok = member["hash"]
	if !ok {
		t.Fatal("Expected to find hash found none")
	}
	hash = hashi.(string)
	if hash != "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff" {
		t.Fatalf("Expected hash to be: 4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff found: %v", hash)
	}
	validateCursor(cursor, t)

	t.Log("Removing edge")
	cursor = "cursor7"
	err = doccache.MutateEdge(&ChainEdge{
		Name: "member",
		From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
		To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	}, true, cursor)
	if err != nil {
		t.Fatalf("MutateEdge for removing failed: %v", err)
	}

	docAsMap, err = doccache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	if err != nil {
		t.Fatalf("GetByHashAsMap failed: %v", err)
	}
	t.Logf("Doc as map: %v", docAsMap)
	if docAsMap == nil {
		t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	}
	membersi, ok = docAsMap["member"]
	if ok {
		t.Fatal("Expected not to find member edge")
	}
	validateCursor(cursor, t)

	t.Log("Removing edge 2")
	cursor = "cursor8"
	err = doccache.MutateEdge(&ChainEdge{
		Name: "member",
		From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
		To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	}, true, cursor)
	if err != nil {
		t.Fatalf("MutateEdge for removing failed: %v", err)
	}

	docAsMap, err = doccache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	if err != nil {
		t.Fatalf("GetByHashAsMap failed: %v", err)
	}
	t.Logf("Doc as map: %v", docAsMap)
	if docAsMap == nil {
		t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	}
	membersi, ok = docAsMap["member"]
	if ok {
		t.Fatal("Expected not to find member edge")
	}
	validateCursor(cursor, t)

}

func validateCursor(cursor string, t *testing.T) {
	if doccache.Cursor.Cursor != cursor {
		t.Fatalf("Expected in memory cursor to be %v found: %v", cursor, doccache.Cursor.Cursor)
	}
	c, err := doccache.getCursor()
	if err != nil {
		t.Fatalf("GetCursor failed: %v", err)
	}
	if c.Cursor != cursor {
		t.Fatalf("Expected in db cursor to be %v found: %v", cursor, c.Cursor)
	}
}

func compareDocs(expected, actual *Document, t *testing.T) {
	if expected.Hash != actual.Hash {
		t.Fatalf("Doc Hashes do not match, expected: %v\n\n found: %v", expected.Hash, actual.Hash)
	}
	if *expected.CreatedDate != *actual.CreatedDate {
		t.Fatalf("Doc CreatedDates do not match, expected: %v\n\n found: %v", expected.CreatedDate, actual.CreatedDate)
	}
	if expected.Creator != actual.Creator {
		t.Fatalf("Doc Creators do not match, expected: %v\n\n found: %v", expected.Creator, actual.Creator)
	}
	if !reflect.DeepEqual(expected.DType, actual.DType) {
		t.Fatalf("Doc DTypes do not match, expected: %v\n\n found: %v", expected.DType, actual.DType)
	}
	if len(expected.ContentGroups) != len(actual.ContentGroups) {
		t.Fatalf("ContentGroups length do not match, expected: %v\n\n found: %v", len(expected.ContentGroups), len(actual.ContentGroups))
	}
	for i, expectedContentGroup := range expected.ContentGroups {
		compareContentGroup(expectedContentGroup, actual.ContentGroups[i], t)
	}
	if len(expected.Certificates) != len(actual.Certificates) {
		t.Fatalf("Certificates length do not match, expected: %v\n\n found: %v", len(expected.Certificates), len(actual.Certificates))
	}
	for i, expectedCertificate := range expected.Certificates {
		compareCertificate(expectedCertificate, actual.Certificates[i], t)
	}
}

func compareContentGroup(expected, actual *ContentGroup, t *testing.T) {
	if expected.ContentGroupSequence != actual.ContentGroupSequence {
		t.Fatalf("ContentGroup ContentGroupSequences do not match, expected: %v\n\n found: %v", expected.ContentGroupSequence, actual.ContentGroupSequence)
	}
	if !reflect.DeepEqual(expected.DType, actual.DType) {
		t.Fatalf("ContentGroup DTypes do not match, expected: %v\n\n found: %v, expected: %v\n\n found: %v", expected.DType, actual.DType, expected, actual)
	}
	if len(expected.Contents) != len(actual.Contents) {
		t.Fatalf("ContentGroup Contents length do not match, expected: %v\n\n found: %v, expected: %v\n\n found: %v", len(expected.Contents), len(actual.Contents), expected, actual)
	}
	for i, expectedContent := range expected.Contents {
		compareContent(expectedContent, actual.Contents[i], expected.ContentGroupSequence, t)
	}
}

func compareContent(expected, actual *Content, contentGroupSequence int, t *testing.T) {
	if expected.Label != actual.Label {
		t.Fatalf("Content Labeles do not match, expected: %v\n\n found: %v, expected: %v\n\n found: %v, contentGroupSequence: %v", expected.Label, actual.Label, expected, actual, contentGroupSequence)
	}
	if expected.Type != actual.Type {
		t.Fatalf("Content Types do not match, expected: %v\n\n found: %v, expected: %v\n\n found: %v, contentGroupSequence: %v", expected.Type, actual.Type, expected, actual, contentGroupSequence)
	}
	if expected.Value != actual.Value {
		t.Fatalf("Content Values do not match, expected: %v\n\n found: %v, expected: %v\n\n found: %v, contentGroupSequence: %v", expected.Value, actual.Value, expected, actual, contentGroupSequence)
	}

	if !reflect.DeepEqual(expected.TimeValue, actual.TimeValue) {
		t.Fatalf("Content Time Values do not match, expected: %v, found: %v, expected: %v, found: %v, contentGroupSequence: %v", expected.TimeValue, actual.TimeValue, expected, actual, contentGroupSequence)
	}

	// if reflect.DeepEqual(expected.IntValue, actual.IntValue) {
	// 	t.Fatalf("Content Int Values do not match, expected: %v, found: %v, expected: %v, found: %v, contentGroupSequence: %v", expected.TimeValue, actual.TimeValue, expected, actual, contentGroupSequence)
	// }

	if expected.ContentSequence != actual.ContentSequence {
		t.Fatalf("Content ContentSequences do not match, expected: %v\n\n, found: %v\n\n, expected: %v\n\n, found: %v\n\n, contentGroupSequence: %v", expected.ContentSequence, actual.ContentSequence, expected, actual, contentGroupSequence)
	}
	if (len(expected.Document) != len(actual.Document)) ||
		(len(expected.Document) == 1 && expected.Document[0].Hash != actual.Document[0].Hash) {
		t.Fatalf("Content Documents do not match, expected: %v\n\n found: %v\n\n expected: %v\n\n found: %v\n\n contentGroupSequence: %v", expected.Document, actual.Document, expected, actual, contentGroupSequence)
	}
	if !reflect.DeepEqual(expected.DType, actual.DType) {
		t.Fatalf("Content DTypes do not match, expected: %v\n\n found: %v\n\n expected: %v\n\n found: %v\n\n contentGroupSequence: %v", expected.DType, actual.DType, expected, actual, contentGroupSequence)
	}
}

func compareCertificate(expected, actual *Certificate, t *testing.T) {
	if expected.Certifier != actual.Certifier {
		t.Fatalf("Certificate Certifiers do not match, expected: %v\n\n found: %v", expected.Certifier, actual.Certifier)
	}
	if expected.Notes != actual.Notes {
		t.Fatalf("Certificate Notess do not match, expected: %v\n\n found: %v", expected.Notes, actual.Notes)
	}
	if *expected.CertificationDate != *actual.CertificationDate {
		t.Fatalf("Certificate CertificationDates do not match, expected: %v\n\n found: %v", expected.CertificationDate, actual.CertificationDate)
	}
	if expected.CertificationSequence != actual.CertificationSequence {
		t.Fatalf("Certificate CertificationSequences do not match, expected: %v\n\n found: %v", expected.CertificationSequence, actual.CertificationSequence)
	}
	if !reflect.DeepEqual(expected.DType, actual.DType) {
		t.Fatalf("Certificate DTypes do not match, expected: %v\n\n found: %v, expected: %v\n\n found: %v", expected.DType, actual.DType, expected, actual)
	}
}

func TestChainDocUnmarshall(t *testing.T) {
	chainDocJSON := `{"certificates":[],"content_groups":[[{"label":"content_group_label","value":["string","settings"]},{"label":"root_node","value":["string","52a7ff82bd6f53b31285e97d6806d886eefb650e79754784e9d923d3df347c91"]},{"label":"paused","value":["int64",0]},{"label":"updated_date","value":["time_point","2021-01-11T21:52:32"]},{"label":"seeds_token_contract","value":["name","token.seeds"]},{"label":"voting_duration_sec","value":["int64",3600]},{"label":"seeds_deferral_factor_x100","value":["int64",100]},{"label":"telos_decide_contract","value":["name","trailservice"]},{"label":"husd_token_contract","value":["name","husd.hypha"]},{"label":"hypha_token_contract","value":["name","token.hypha"]},{"label":"seeds_escrow_contract","value":["name","escrow.seeds"]},{"label":"publisher_contract","value":["name","publsh.hypha"]},{"label":"treasury_contract","value":["name","bank.hypha"]},{"label":"last_ballot_id","value":["name","hypha1....1cf"]},{"label":"hypha_deferral_factor_x100","value":["int64",25]},{"label":"client_version","value":["string","0.2.0 pre-release"]},{"label":"contract_version","value":["string","0.2.0 pre-release"]}],[{"label":"content_group_label","value":["string","system"]},{"label":"type","value":["name","settings"]},{"label":"node_label","value":["string","Settings"]}]],"contract":"dao.hypha","created_date":"2021-01-11T21:52:32","creator":"dao.hypha","hash":"3e06f9f93fb27ad04a2e97dfce9796c2d51b73721d6270e1c0ea6bf7e79c944b","id":4957}`
	chainDoc := &ChainDocument{}
	err := json.Unmarshal([]byte(chainDocJSON), chainDoc)
	if err != nil {
		t.Fatalf("Unmarshalling failed: %v", err)
	}
}

func TestChainEdgeUnmarshall(t *testing.T) {
	chainDocEdge := `{"contract":"dao.hypha","created_date":"2021-01-11T21:52:32","creator":"dao.hypha","edge_name":"settings","from_node":"52a7ff82bd6f53b31285e97d6806d886eefb650e79754784e9d923d3df347c91","from_node_edge_name_index":493623357,"from_node_to_node_index":340709097,"id":2475211255,"to_node":"3e06f9f93fb27ad04a2e97dfce9796c2d51b73721d6270e1c0ea6bf7e79c944b","to_node_edge_name_index":2119673673}`
	chainEdge := &ChainEdge{}
	err := json.Unmarshal([]byte(chainDocEdge), chainEdge)
	if err != nil {
		t.Fatalf("Unmarshalling failed: %v", err)
	}
}
