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
	tutil "github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/util"
	"gotest.tools/assert"
)

var dg *dgraph.Dgraph
var cache *doccache.Doccache

var memberType = &gql.SimplifiedType{
	Name: "Member",
	Fields: map[string]*gql.SimplifiedField{
		"details_account_n": {
			Name:  "details_account_n",
			Type:  "String",
			Index: "exact",
		},
	},
	ExtendsDocument: true,
}

var periodType = &gql.SimplifiedType{
	Name: "Period",
	Fields: map[string]*gql.SimplifiedField{
		"details_number_i": {
			Name:  "details_number_i",
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
	config, err := util.LoadConfig("./config.yml")
	if err != nil {
		log.Fatal(err, "Failed to load configuration")
	}
	admin := gql.NewAdmin(config.GQLAdminURL)
	client := gql.NewClient(config.GQLClientURL)
	dg, err = dgraph.New("")
	if err != nil {
		log.Fatal(err, "Unable to create dgraph")
	}
	err = dg.DropAll()
	if err != nil {
		log.Fatal(err, "Unable to drop all")
	}
	time.Sleep(time.Second * 2)
	cache, err = doccache.New(dg, admin, client, config.TypeMappings, nil)
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
					Label: "str_to_int",
					Value: []interface{}{
						"string",
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
			"details_rootNode_n": {
				Name:  "details_rootNode_n",
				Type:  "String",
				Index: "exact",
			},
			"details_hvoiceSalaryPerPhase_a": {
				Name:  "details_hvoiceSalaryPerPhase_a",
				Type:  "String",
				Index: "term",
			},
			"details_timeShareX100_i": {
				Name:  "details_timeShareX100_i",
				Type:  "Int64",
				Index: "int64",
			},
			"details_strToInt_s": {
				Name:  "details_strToInt_s",
				Type:  "String",
				Index: "regexp",
			},
			"details_startPeriod_c": {
				Name:  "details_startPeriod_c",
				Type:  "String",
				Index: "exact",
			},
			"details_startPeriod_c_edge": {
				Name: "details_startPeriod_c_edge",
				Type: "Period",
			},
			"system_originalApprovedDate_t": {
				Name:  "system_originalApprovedDate_t",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
		},
		ExtendsDocument: true,
	}
	expectedDHOInstance := &gql.SimplifiedInstance{
		SimplifiedType: expectedDhoType,
		Values: map[string]interface{}{
			"hash":                           dhoHash,
			"createdDate":                    "2020-11-12T18:27:47.000Z",
			"creator":                        "dao.hypha",
			"type":                           "Dho",
			"details_rootNode_n":             "dao.hypha",
			"details_hvoiceSalaryPerPhase_a": "4133.04 HVOICE",
			"details_timeShareX100_i":        int64(60),
			"details_strToInt_s":             "60",
			"details_startPeriod_c":          period1Hash,
			"details_startPeriod_c_edge":     doccache.GetEdgeValue(period1Hash),
			"system_originalApprovedDate_t":  "2021-04-12T05:09:36.5Z",
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
		To:   member1Hash,
	}, false, cursor)
	assert.NilError(t, err)

	expectedDhoType.SetField("member", &gql.SimplifiedField{
		Name:    "member",
		Type:    "Document",
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
					Label: "str_to_int",
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
		"details_strToInt_i",
		&gql.SimplifiedField{
			Name:  "details_strToInt_i",
			Type:  "Int64",
			Index: "int64",
		},
	)
	expectedDhoType.SetField(
		"details_periodCount_i",
		&gql.SimplifiedField{
			Name:  "details_periodCount_i",
			Type:  "Int64",
			Index: "int64",
		},
	)
	expectedDhoType.SetField(
		"system_endPeriod_c",
		&gql.SimplifiedField{
			Name:  "system_endPeriod_c",
			Type:  "String",
			Index: "exact",
		},
	)
	expectedDhoType.SetField(
		"system_endPeriod_c_edge",
		&gql.SimplifiedField{
			Name: "system_endPeriod_c_edge",
			Type: "Period",
		},
	)
	expectedDHOInstance.SetValue("details_periodCount_i", int64(50))
	expectedDHOInstance.SetValue("details_timeShareX100_i", nil)
	expectedDHOInstance.SetValue("details_strToInt_s", nil)
	expectedDHOInstance.SetValue("details_strToInt_i", int64(60))
	expectedDHOInstance.SetValue("system_originalApprovedDate_t", "2021-05-12T05:09:36.5Z")
	expectedDHOInstance.SetValue("details_hvoiceSalaryPerPhase_a", "4233.04 HVOICE")
	expectedDHOInstance.SetValue("system_endPeriod_c", period2Hash)
	expectedDHOInstance.SetValue("system_endPeriod_c_edge", doccache.GetEdgeValue(period2Hash))

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
					Label: "str_to_int",
					Value: []interface{}{
						"int64",
						"60",
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

	expectedDHOInstance.SetValue("details_startPeriod_c", nil)
	expectedDHOInstance.SetValue("details_startPeriod_c_edge", nil)

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

}

func TestDocumentCreationDeduceType(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	hash := "d4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		Hash:        hash,
		CreatedDate: createdDate,
		Creator:     "dao.hypha",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"pass",
					},
				},
				{
					Label: "vote_power",
					Value: []interface{}{
						"asset",
						"0.00 HVOICE",
					},
				},
			},
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"fail",
					},
				},
				{
					Label: "vote_power",
					Value: []interface{}{
						"asset",
						"1.00 HVOICE",
					},
				},
			},
		},
	}

	expectedInstance := &gql.SimplifiedInstance{
		SimplifiedType: &gql.SimplifiedType{
			Name: "VoteTally",
			Fields: map[string]*gql.SimplifiedField{
				"pass_votePower_a": {
					Name:  "pass_votePower_a",
					Type:  "String",
					Index: "term",
				},
				"fail_votePower_a": {
					Name:  "fail_votePower_a",
					Type:  "String",
					Index: "term",
				},
			},
			ExtendsDocument: true,
		},
		Values: map[string]interface{}{
			"hash":             hash,
			"createdDate":      "2020-11-12T18:27:47.000Z",
			"creator":          "dao.hypha",
			"type":             "VoteTally",
			"pass_votePower_a": "0.00 HVOICE",
			"fail_votePower_a": "1.00 HVOICE",
		},
	}

	cursor := "cursor0"
	err := cache.StoreDocument(chainDoc1, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance)
	assertCursor(t, cursor)

	cursor = "cursor1"

	err = cache.DeleteDocument(chainDoc1, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, hash, "VoteTally")
	assertCursor(t, cursor)

}

func TestMissingCoreEdge(t *testing.T) {

	t.Log("Store assignment 1 with related core edge non existant")
	createdDate := "2020-11-12T18:27:47.000"
	period1Hash := "a5ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	hash := "b5ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment1 := &domain.ChainDocument{
		ID:          0,
		Hash:        hash,
		CreatedDate: createdDate,
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
						"assignment",
					},
				},
			},
		},
	}

	expectedType := &gql.SimplifiedType{
		Name: "Assignment",
		Fields: map[string]*gql.SimplifiedField{
			"details_startPeriod_c": {
				Name:  "details_startPeriod_c",
				Type:  "String",
				Index: "exact",
			},
		},
		ExtendsDocument: true,
	}

	expectedInstance := &gql.SimplifiedInstance{
		SimplifiedType: expectedType,
		Values: map[string]interface{}{
			"hash":                  hash,
			"createdDate":           "2020-11-12T18:27:47.000Z",
			"creator":               "dao.hypha",
			"type":                  "Assignment",
			"details_startPeriod_c": period1Hash,
		},
	}

	cursor := "cursor0"
	err := cache.StoreDocument(assignment1, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance)
	assertCursor(t, cursor)

	t.Log("Store core edge")
	period1Doc := getPeriodDoc(period1Hash, 1)
	period1Instance := getPeriodInstance(period1Hash, 1)
	cursor = "cursor1"
	err = cache.StoreDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, period1Instance)
	assertCursor(t, cursor)

	t.Log("Store assignment 2 with related core edge")
	hash2 := "c5ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	assignment2 := &domain.ChainDocument{
		ID:          0,
		Hash:        hash2,
		CreatedDate: createdDate,
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
						"assignment",
					},
				},
			},
		},
	}

	expectedType.SetField("details_startPeriod_c_edge",
		&gql.SimplifiedField{
			Name: "details_startPeriod_c_edge",
			Type: "Period",
		})

	expectedInstance2 := &gql.SimplifiedInstance{
		SimplifiedType: expectedType,
		Values: map[string]interface{}{
			"hash":                       hash2,
			"createdDate":                "2020-11-12T18:27:47.000Z",
			"creator":                    "dao.hypha",
			"type":                       "Assignment",
			"details_startPeriod_c":      period1Hash,
			"details_startPeriod_c_edge": map[string]interface{}{"hash": period1Hash},
		},
	}

	cursor = "cursor2"
	err = cache.StoreDocument(assignment2, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance2)
	assertCursor(t, cursor)

	t.Log("Verify assignment 1 has a nil core edge")
	expectedInstance.SetValue("details_startPeriod_c_edge", nil)
	assertInstance(t, expectedInstance)

	cursor = "cursor4"

	t.Log("Delete core edge document")
	err = cache.DeleteDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, period1Hash, "Period")
	assertCursor(t, cursor)

	t.Log("Verify assignment 2 has a nil core edge")
	expectedInstance2.SetValue("details_startPeriod_c_edge", nil)
	assertInstance(t, expectedInstance2)

	t.Log("Store core edge again")
	cursor = "cursor5"
	err = cache.StoreDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, period1Instance)
	assertCursor(t, cursor)

	expectedInstance2.SetValue("details_startPeriod_c_edge", map[string]interface{}{"hash": period1Hash})

	t.Log("Update assignment 2, should relink core edge")
	cursor = "cursor6"
	err = cache.StoreDocument(assignment2, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance2)
	assertCursor(t, cursor)

	t.Log("Delete documents")
	cursor = "cursor7"
	err = cache.DeleteDocument(assignment1, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, hash, "Assignment")
	assertCursor(t, cursor)

	cursor = "cursor8"
	err = cache.DeleteDocument(assignment2, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, hash2, "Assignment")
	assertCursor(t, cursor)

	cursor = "cursor9"
	err = cache.DeleteDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, period1Hash, "Period")
	assertCursor(t, cursor)

}

func assertCursor(t *testing.T, cursor string) {
	expected := gql.NewCursorInstance(doccache.CursorId, cursor)
	actual, err := cache.GetInstance(doccache.CursorId, gql.CursorSimplifiedType, nil)
	assert.NilError(t, err)
	tutil.AssertSimplifiedInstance(t, actual, expected)
}

func assertInstance(t *testing.T, expected *gql.SimplifiedInstance) {
	actualType, err := cache.Schema.GetSimplifiedType(expected.SimplifiedType.Name)
	assert.NilError(t, err)
	actual, err := cache.GetInstance(expected.GetValue("hash"), actualType, nil)
	assert.NilError(t, err)
	fmt.Println("Expected: ", expected)
	fmt.Println("Actual: ", actual)
	tutil.AssertSimplifiedInstance(t, actual, expected)
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
			"hash":              hash,
			"createdDate":       "2020-11-12T19:27:47.000Z",
			"creator":           account,
			"type":              "Member",
			"details_account_n": account,
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
			"hash":             hash,
			"createdDate":      "2020-11-12T18:27:47.000Z",
			"creator":          "dao.hypha",
			"type":             "Period",
			"details_number_i": number,
		},
	}
}
