package doccache_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

var dg *dgraph.Dgraph
var cache *doccache.Doccache

var memberType = &gql.SimplifiedType{
	Name: "Member",
	Fields: map[string]*gql.SimplifiedField{
		"details_account": {
			Name:  "details_account",
			Type:  "String",
			Index: "exact",
		},
	},
	ExtendsDocument: true,
}

var periodType = &gql.SimplifiedType{
	Name: "Period",
	Fields: map[string]*gql.SimplifiedField{
		"details_number": {
			Name:  "details_number",
			Type:  "Int64",
			Index: "int64",
		},
	},
	ExtendsDocument: true,
}

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
	admin := gql.NewAdmin("http://localhost:8080/admin")
	client := gql.NewClient("http://localhost:8080/graphql")
	dg, err = dgraph.New("")
	if err != nil {
		log.Fatal(err, "Unable to create dgraph")
	}
	err = dg.DropAll()
	if err != nil {
		log.Fatal(err, "Unable to drop all")
	}
	time.Sleep(time.Second * 2)
	cache, err = doccache.New(dg, admin, client, nil)
	if err != nil {
		log.Fatal(err, "Failed creating DocCache")
	}
}

func afterAll() {
	dg.Close()
}

func TestOpCycle(t *testing.T) {
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorId)

	t.Logf("Storing period 1 document")
	period1Hash := "h4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	periodDoc := getPeriodDoc(period1Hash, 1)
	expectedPeriodInstance := getPeriodInstance(period1Hash, 1)

	cursor := "cursor0"
	err := cache.StoreDocument(periodDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedPeriodInstance)
	assertCursor(t, cursor)

	t.Logf("Storing dho document")
	dhoHash := "z4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	dhoDoc := &domain.ChainDocument{
		ID:          0,
		Hash:        dhoHash,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "root_node",
					Value: []interface{}{
						"name",
						"dao.hypha",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
					},
				},
				{
					Label: "hvoice_salary_per_phase",
					Value: []interface{}{
						"asset",
						"4133.04 HVOICE",
					},
				},
				{
					Label: "time_share_x100",
					Value: []interface{}{
						"int64",
						"60",
					},
				},
				{
					Label: "start_period",
					Value: []interface{}{
						"checksum256",
						period1Hash,
					},
				},
			},
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "type",
					Value: []interface{}{
						"name",
						"dho",
					},
				},
				{
					Label: "original_approved_date",
					Value: []interface{}{
						"time_point",
						"2021-04-12T05:09:36.5",
					},
				},
			},
		},
	}
	expectedDhoType := &gql.SimplifiedType{
		Name: "Dho",
		Fields: map[string]*gql.SimplifiedField{
			"details_rootNode": {
				Name:  "details_rootNode",
				Type:  "String",
				Index: "exact",
			},
			"details_hvoiceSalaryPerPhase": {
				Name:  "details_hvoiceSalaryPerPhase",
				Type:  "String",
				Index: "term",
			},
			"details_timeShareX100": {
				Name:  "details_timeShareX100",
				Type:  "Int64",
				Index: "int64",
			},
			"details_startPeriod": {
				Name:  "details_startPeriod",
				Type:  "String",
				Index: "exact",
			},
			"details_startPeriod_edge": {
				Name: "details_startPeriod_edge",
				Type: "Period",
			},
			"system_originalApprovedDate": {
				Name:  "system_originalApprovedDate",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
		},
		ExtendsDocument: true,
	}
	expectedDHOInstance := &gql.SimplifiedInstance{
		SimplifiedType: expectedDhoType,
		Values: map[string]interface{}{
			"hash":                         dhoHash,
			"createdDate":                  "2020-11-12T18:27:47.000Z",
			"creator":                      "dao.hypha",
			"type":                         "Dho",
			"details_rootNode":             "dao.hypha",
			"details_hvoiceSalaryPerPhase": "4133.04 HVOICE",
			"details_timeShareX100":        int64(60),
			"details_startPeriod":          period1Hash,
			"details_startPeriod_edge":     doccache.GetEdgeValue(period1Hash),
			"system_originalApprovedDate":  "2021-04-12T05:09:36.5Z",
		},
	}
	cursor = "cursor1"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing member document")
	member1Hash := "a4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	memberDoc := getMemberDoc(member1Hash, "member1")
	expectedMemberInstance := getMemberInstance(member1Hash, "member1")
	cursor = "cursor2"

	err = cache.StoreDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedMemberInstance)
	assertCursor(t, cursor)

	t.Logf("Storing another member document")
	member2Hash := "b4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	memberDoc = getMemberDoc(member2Hash, "member2")
	expectedMemberInstance = getMemberInstance(member2Hash, "member2")
	cursor = "cursor3"

	err = cache.StoreDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedMemberInstance)
	assertCursor(t, cursor)

	t.Log("Adding edge")
	cursor = "cursor4"
	err = cache.MutateEdge(&domain.ChainEdge{
		Name: "member",
		From: dhoHash,
		To:   member1Hash + "1",
	}, false, cursor)
	assert.NilError(t, err)

	expectedDhoType.SetField("member", &gql.SimplifiedField{
		Name:    "member",
		Type:    "Member",
		IsArray: true,
		NonNull: false,
	})
	expectedMemberEdge := []map[string]interface{}{
		{"hash": member1Hash},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Adding second edge")
	cursor = "cursor5"
	err = cache.MutateEdge(&domain.ChainEdge{
		Name: "member",
		From: dhoHash,
		To:   member2Hash,
	}, false, cursor)
	assert.NilError(t, err)

	expectedMemberEdge = []map[string]interface{}{
		{"hash": member1Hash},
		{"hash": member2Hash},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing period 2 document")
	period2Hash := "i4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	periodDoc = getPeriodDoc(period2Hash, 2)
	expectedPeriodInstance = getPeriodInstance(period2Hash, 2)

	cursor = "cursorA"
	err = cache.StoreDocument(periodDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedPeriodInstance)
	assertCursor(t, cursor)

	t.Log("Update DHO document: Update values, add coreedge, remove core field")
	dhoDoc = &domain.ChainDocument{
		ID:          0,
		Hash:        dhoHash,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "root_node",
					Value: []interface{}{
						"name",
						"dao.hypha",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
					},
				},
				{
					Label: "hvoice_salary_per_phase",
					Value: []interface{}{
						"asset",
						"4233.04 HVOICE",
					},
				},
				{
					Label: "start_period",
					Value: []interface{}{
						"checksum256",
						period1Hash,
					},
				},
				{
					Label: "period_count",
					Value: []interface{}{
						"int64",
						"50",
					},
				},
			},
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "type",
					Value: []interface{}{
						"name",
						"dho",
					},
				},
				{
					Label: "original_approved_date",
					Value: []interface{}{
						"time_point",
						"2021-05-12T05:09:36.5",
					},
				},
				{
					Label: "end_period",
					Value: []interface{}{
						"checksum256",
						period2Hash,
					},
				},
			},
		},
	}

	expectedDhoType.SetField(
		"details_periodCount",
		&gql.SimplifiedField{
			Name:  "details_periodCount",
			Type:  "Int64",
			Index: "int64",
		},
	)
	expectedDhoType.SetField(
		"system_endPeriod",
		&gql.SimplifiedField{
			Name:  "system_endPeriod",
			Type:  "String",
			Index: "exact",
		},
	)
	expectedDhoType.SetField(
		"system_endPeriod_edge",
		&gql.SimplifiedField{
			Name: "system_endPeriod_edge",
			Type: "Period",
		},
	)
	expectedDHOInstance.SetValue("details_periodCount", int64(50))
	expectedDHOInstance.SetValue("details_timeShareX100", nil)
	expectedDHOInstance.SetValue("system_originalApprovedDate", "2021-05-12T05:09:36.5Z")
	expectedDHOInstance.SetValue("details_hvoiceSalaryPerPhase", "4233.04 HVOICE")
	expectedDHOInstance.SetValue("system_endPeriod", period2Hash)
	expectedDHOInstance.SetValue("system_endPeriod_edge", doccache.GetEdgeValue(period2Hash))

	cursor = "cursor6"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Deleting edge")
	cursor = "cursor7"
	err = cache.MutateEdge(&domain.ChainEdge{
		Name: "member",
		From: dhoHash,
		To:   member1Hash,
	}, true, cursor)
	assert.NilError(t, err)

	expectedDhoType.SetField("member", &gql.SimplifiedField{
		Name:    "member",
		Type:    "Member",
		IsArray: true,
		NonNull: false,
	})
	expectedMemberEdge = []map[string]interface{}{
		{"hash": member2Hash},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)

	t.Log("Update 2 DHO document: remove core edge")
	dhoDoc = &domain.ChainDocument{
		ID:          0,
		Hash:        dhoHash,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "root_node",
					Value: []interface{}{
						"name",
						"dao.hypha",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
					},
				},
				{
					Label: "hvoice_salary_per_phase",
					Value: []interface{}{
						"asset",
						"4233.04 HVOICE",
					},
				},
				{
					Label: "period_count",
					Value: []interface{}{
						"int64",
						"50",
					},
				},
			},
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "type",
					Value: []interface{}{
						"name",
						"dho",
					},
				},
				{
					Label: "original_approved_date",
					Value: []interface{}{
						"time_point",
						"2021-05-12T05:09:36.5",
					},
				},
				{
					Label: "end_period",
					Value: []interface{}{
						"checksum256",
						period2Hash,
					},
				},
			},
		},
	}

	expectedDHOInstance.SetValue("details_startPeriod", nil)
	expectedDHOInstance.SetValue("details_startPeriod_edge", nil)

	cursor = "cursorB"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Deleting second edge")
	cursor = "cursor8"
	err = cache.MutateEdge(&domain.ChainEdge{
		Name: "member",
		From: dhoHash,
		To:   member2Hash,
	}, true, cursor)
	assert.NilError(t, err)

	expectedDhoType.SetField("member", &gql.SimplifiedField{
		Name:    "member",
		Type:    "Member",
		IsArray: true,
		NonNull: false,
	})
	expectedMemberEdge = []map[string]interface{}{}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)

	t.Logf("Deleting member1 document")
	memberDoc = getMemberDoc(member1Hash, "member1")
	cursor = "cursor9"

	err = cache.DeleteDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, member1Hash, "Member")
	assertCursor(t, cursor)

	t.Logf("Deleting member2 document")
	memberDoc = getMemberDoc(member2Hash, "member2")
	cursor = "cursor10"

	err = cache.DeleteDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, member2Hash, "Member")
	assertCursor(t, cursor)

	t.Logf("Deleting dho document")
	cursor = "cursor11"
	err = cache.DeleteDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, dhoHash, "Dho")
	assertCursor(t, cursor)
	// t.Logf("Updating document certificates")

	// certificationDate := "2020-11-12T20:27:47.000"
	// chainDoc1.Certificates = []*domain.ChainCertificate{
	// 	{
	// 		Certifier:         "sebastian",
	// 		Notes:             "Sebastian's Notes",
	// 		CertificationDate: certificationDate,
	// 	},
	// }
	// expectedDoc1.Certificates = []*domain.Certificate{
	// 	{
	// 		Certifier:             "sebastian",
	// 		Notes:                 "Sebastian's Notes",
	// 		CertificationDate:     domain.ToTime(certificationDate),
	// 		CertificationSequence: 1,
	// 		DType:                 []string{"Certificate"},
	// 	},
	// }
	// cursor = "cursor2"
	// err = cache.StoreDocument(chainDoc1, cursor)
	// if err != nil {
	// 	t.Fatalf("StoreDocument failed: %v", err)
	// }

	// doc, err = cache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	// if err != nil {
	// 	t.Fatalf("GetByHash failed: %v", err)
	// }
	// compareDocs(expectedDoc1, doc, t)
	// validateCursor(cursor, t)

	// t.Logf("Updating document certificates 2")

	// certificationDate = "2020-11-14T20:27:47.000"
	// chainDoc1.Certificates = append(chainDoc1.Certificates, &domain.ChainCertificate{
	// 	Certifier:         "pedro",
	// 	Notes:             "Pedro's Notes",
	// 	CertificationDate: certificationDate,
	// })
	// expectedDoc1.Certificates = append(expectedDoc1.Certificates, &domain.Certificate{
	// 	Certifier:             "pedro",
	// 	Notes:                 "Pedro's Notes",
	// 	CertificationDate:     domain.ToTime(certificationDate),
	// 	CertificationSequence: 2,
	// 	DType:                 []string{"Certificate"},
	// })

	// cursor = "cursor3"
	// err = cache.StoreDocument(chainDoc1, cursor)
	// if err != nil {
	// 	t.Fatalf("StoreDocument failed: %v", err)
	// }

	// doc, err = cache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	// if err != nil {
	// 	t.Fatalf("GetByHash failed: %v", err)
	// }
	// compareDocs(expectedDoc1, doc, t)
	// validateCursor(cursor, t)

	// createdDate = "2020-11-12T22:09:12.000"
	// startTime := "2021-04-01T15:50:54.291"
	// chainDoc2 := &domain.ChainDocument{
	// 	ID:          1,
	// 	Hash:        "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	// 	CreatedDate: createdDate,
	// 	Creator:     "dao.hypha1",
	// 	ContentGroups: [][]*domain.ChainContent{
	// 		{
	// 			{
	// 				Label: "member",
	// 				Value: []interface{}{
	// 					"name",
	// 					"1onefiftyfor",
	// 				},
	// 			},
	// 			{
	// 				Label: "role",
	// 				Value: []interface{}{
	// 					"name",
	// 					"dev",
	// 				},
	// 			},
	// 			{
	// 				Label: "start_time",
	// 				Value: []interface{}{
	// 					"time_point",
	// 					startTime,
	// 				},
	// 			},
	// 		},
	// 		{
	// 			{
	// 				Label: "root",
	// 				Value: []interface{}{
	// 					"checksum256",
	// 					"d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
	// 				},
	// 			},
	// 			{
	// 				Label: "vote_count",
	// 				Value: []interface{}{
	// 					"int64",
	// 					89,
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// voteCount := int64(89)
	// expectedDoc2 := &domain.Document{
	// 	Hash:        "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	// 	CreatedDate: domain.ToTime(createdDate),
	// 	Creator:     "dao.hypha1",
	// 	DType:       []string{"Document"},
	// 	ContentGroups: []*domain.ContentGroup{
	// 		{
	// 			ContentGroupSequence: 1,
	// 			DType:                []string{"ContentGroup"},
	// 			Contents: []*domain.Content{
	// 				{
	// 					Label:           "member",
	// 					Type:            "name",
	// 					Value:           "1onefiftyfor",
	// 					ContentSequence: 1,
	// 					DType:           []string{"Content"},
	// 				},
	// 				{
	// 					Label:           "role",
	// 					Type:            "name",
	// 					Value:           "dev",
	// 					ContentSequence: 2,
	// 					DType:           []string{"Content"},
	// 				},
	// 				{
	// 					Label:           "start_time",
	// 					Type:            "time_point",
	// 					Value:           startTime,
	// 					TimeValue:       domain.ToTime(startTime),
	// 					ContentSequence: 3,
	// 					DType:           []string{"Content"},
	// 				},
	// 			},
	// 		},
	// 		{
	// 			ContentGroupSequence: 2,
	// 			DType:                []string{"ContentGroup"},
	// 			Contents: []*domain.Content{
	// 				{
	// 					Label:           "root",
	// 					Type:            "checksum256",
	// 					Value:           "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
	// 					ContentSequence: 1,
	// 					DType:           []string{"Content"},
	// 					Document:        []*domain.Document{expectedDoc1},
	// 				},
	// 				{
	// 					Label:           "vote_count",
	// 					Type:            "int64",
	// 					Value:           "89",
	// 					IntValue:        &voteCount,
	// 					ContentSequence: 2,
	// 					DType:           []string{"Content"},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// cursor = "cursor4"
	// t.Log("Storing another document")
	// err = cache.StoreDocument(chainDoc2, cursor)
	// if err != nil {
	// 	t.Fatalf("StoreDocument failed: %v", err)
	// }

	// doc, err = cache.GetByHash("4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff", &RequestConfig{ContentGroups: true, Certificates: true})
	// if err != nil {
	// 	t.Fatalf("GetByHash failed: %v", err)
	// }
	// compareDocs(expectedDoc2, doc, t)
	// validateCursor(cursor, t)

	// t.Log("Check original document wasn't modified")
	// doc, err = cache.GetByHash("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true})
	// if err != nil {
	// 	t.Fatalf("GetByHash failed: %v", err)
	// }
	// compareDocs(expectedDoc1, doc, t)

	// t.Log("Adding edge")
	// cursor = "cursor5"
	// err = cache.MutateEdge(&domain.ChainEdge{
	// 	Name: "member",
	// 	From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
	// 	To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	// }, false, cursor)
	// if err != nil {
	// 	t.Fatalf("MutateEdge for adding failed: %v", err)
	// }

	// docAsMap, err := cache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	// if err != nil {
	// 	t.Fatalf("GetByHashAsMap failed: %v", err)
	// }
	// t.Logf("Doc as map: %v", docAsMap)
	// if docAsMap == nil {
	// 	t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	// }
	// membersi, ok := docAsMap["member"]
	// if !ok {
	// 	t.Fatal("Expected to find member edge found none")
	// }
	// members := membersi.([]interface{})
	// if len(members) != 1 {
	// 	t.Fatalf("Expected to find 1 member found: %v", len(members))
	// }
	// member := members[0].(map[string]interface{})
	// hashi, ok := member["hash"]
	// if !ok {
	// 	t.Fatal("Expected to find hash found none")
	// }
	// hash := hashi.(string)
	// if hash != "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff" {
	// 	t.Fatalf("Expected hash to be: 4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff found: %v", hash)
	// }
	// validateCursor(cursor, t)

	// t.Log("Adding same edge")
	// cursor = "cursor6"
	// err = cache.MutateEdge(&domain.ChainEdge{
	// 	Name: "member",
	// 	From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
	// 	To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	// }, false, cursor)
	// if err != nil {
	// 	t.Fatalf("MutateEdge for adding failed: %v", err)
	// }

	// docAsMap, err = cache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	// if err != nil {
	// 	t.Fatalf("GetByHashAsMap failed: %v", err)
	// }
	// t.Logf("Doc as map: %v", docAsMap)
	// if docAsMap == nil {
	// 	t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	// }
	// membersi, ok = docAsMap["member"]
	// if !ok {
	// 	t.Fatal("Expected to find member edge found none")
	// }
	// members = membersi.([]interface{})
	// if len(members) != 1 {
	// 	t.Fatalf("Expected to find 1 member found: %v", len(members))
	// }
	// member = members[0].(map[string]interface{})
	// hashi, ok = member["hash"]
	// if !ok {
	// 	t.Fatal("Expected to find hash found none")
	// }
	// hash = hashi.(string)
	// if hash != "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff" {
	// 	t.Fatalf("Expected hash to be: 4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff found: %v", hash)
	// }
	// validateCursor(cursor, t)

	// t.Log("Removing edge")
	// cursor = "cursor7"
	// err = cache.MutateEdge(&domain.ChainEdge{
	// 	Name: "member",
	// 	From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
	// 	To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	// }, true, cursor)
	// if err != nil {
	// 	t.Fatalf("MutateEdge for removing failed: %v", err)
	// }

	// docAsMap, err = cache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	// if err != nil {
	// 	t.Fatalf("GetByHashAsMap failed: %v", err)
	// }
	// t.Logf("Doc as map: %v", docAsMap)
	// if docAsMap == nil {
	// 	t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	// }
	// membersi, ok = docAsMap["member"]
	// if ok {
	// 	t.Fatal("Expected not to find member edge")
	// }
	// validateCursor(cursor, t)

	// t.Log("Removing edge 2")
	// cursor = "cursor8"
	// err = cache.MutateEdge(&domain.ChainEdge{
	// 	Name: "member",
	// 	From: "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e",
	// 	To:   "4190fc69b4f88f23ae45828a2df64f79bd687a3cdba8c84fa5a89ce9b88de8ff",
	// }, true, cursor)
	// if err != nil {
	// 	t.Fatalf("MutateEdge for removing failed: %v", err)
	// }

	// docAsMap, err = cache.GetByHashAsMap("d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e", &RequestConfig{ContentGroups: true, Certificates: true, Edges: []string{"member"}})
	// if err != nil {
	// 	t.Fatalf("GetByHashAsMap failed: %v", err)
	// }
	// t.Logf("Doc as map: %v", docAsMap)
	// if docAsMap == nil {
	// 	t.Fatal("Expected to find document: d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e, found none")
	// }
	// membersi, ok = docAsMap["member"]
	// if ok {
	// 	t.Fatal("Expected not to find member edge")
	// }
	// validateCursor(cursor, t)

}

func assertCursor(t *testing.T, cursor string) {
	expected := gql.NewCursorInstance(doccache.CursorId, cursor)
	actual, err := cache.GetInstance(doccache.CursorId, gql.CursorSimplifiedType, nil)
	assert.NilError(t, err)
	util.AssertSimplifiedInstance(t, actual, expected)
}

func assertInstance(t *testing.T, expected *gql.SimplifiedInstance) {
	actualType, err := cache.Schema.GetSimplifiedType(expected.SimplifiedType.Name)
	assert.NilError(t, err)
	actual, err := cache.GetInstance(expected.GetValue("hash"), actualType, nil)
	assert.NilError(t, err)
	fmt.Println("Expected: ", expected)
	fmt.Println("Actual: ", actual)
	util.AssertSimplifiedInstance(t, actual, expected)
}

func assertInstanceNotExists(t *testing.T, hash, typeName string) {
	actualType, err := cache.Schema.GetSimplifiedType(typeName)
	assert.NilError(t, err)
	actual, err := cache.GetInstance(hash, actualType, nil)
	assert.NilError(t, err)
	assert.Assert(t, actual == nil)
}

func getMemberDoc(hash, account string) *domain.ChainDocument {
	return &domain.ChainDocument{
		ID:          1,
		Hash:        hash,
		CreatedDate: "2020-11-12T19:27:47.000",
		Creator:     account,
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
					},
				},
				{
					Label: "account",
					Value: []interface{}{
						"name",
						account,
					},
				},
			},
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "type",
					Value: []interface{}{
						"name",
						"member",
					},
				},
			},
		},
	}
}

func getMemberInstance(hash, account string) *gql.SimplifiedInstance {
	return &gql.SimplifiedInstance{
		SimplifiedType: memberType,
		Values: map[string]interface{}{
			"hash":            hash,
			"createdDate":     "2020-11-12T19:27:47.000Z",
			"creator":         account,
			"type":            "Member",
			"details_account": account,
		},
	}
}

func getPeriodDoc(hash string, number int64) *domain.ChainDocument {
	return &domain.ChainDocument{
		ID:          1,
		Hash:        hash,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
					},
				},
				{
					Label: "number",
					Value: []interface{}{
						"int64",
						number,
					},
				},
			},
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "type",
					Value: []interface{}{
						"name",
						"period",
					},
				},
			},
		},
	}
}

func getPeriodInstance(hash string, number int64) *gql.SimplifiedInstance {
	return &gql.SimplifiedInstance{
		SimplifiedType: periodType,
		Values: map[string]interface{}{
			"hash":           hash,
			"createdDate":    "2020-11-12T18:27:47.000Z",
			"creator":        "dao.hypha",
			"type":           "Period",
			"details_number": number,
		},
	}
}
