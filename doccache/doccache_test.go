package doccache_test

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/config"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	tutil "github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

var cfg *config.Config
var dg *dgraph.Dgraph
var cache *doccache.Doccache
var admin *gql.Admin
var client *gql.Client

var userType = gql.NewSimplifiedType(
	"User",
	map[string]*gql.SimplifiedField{
		"details_account_n": {
			Name:    "details_account_n",
			Type:    "String",
			Indexes: gql.NewIndexes("exact", "regexp"),
		},
	},
	gql.DocumentSimplifiedInterface,
)

var memberType = gql.NewSimplifiedType(
	"Member",
	map[string]*gql.SimplifiedField{
		"details_account_n": {
			Name:    "details_account_n",
			Type:    "String",
			Indexes: gql.NewIndexes("exact", "regexp"),
		},
	},
	gql.DocumentSimplifiedInterface,
)

var periodType = gql.NewSimplifiedType(
	"Period",
	map[string]*gql.SimplifiedField{
		"details_number_i": {
			Name:    "details_number_i",
			Type:    "Int64",
			Indexes: gql.NewIndexes("int64"),
		},
	},
	gql.DocumentSimplifiedInterface,
)

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
}

func afterAll() {
	dg.Close()
}

func setUp(configPath string) {
	var err error
	cfg, err = config.LoadConfig(configPath)
	if err != nil {
		log.Fatal(err, "Failed to load configuration")
	}
	admin = gql.NewAdmin(cfg.GQLAdminURL)
	client = gql.NewClient(cfg.GQLClientURL)
	dg, err = dgraph.New("")
	if err != nil {
		log.Fatal(err, "Unable to create dgraph")
	}
	err = dg.DropAll()
	if err != nil {
		log.Fatal(err, "Unable to drop all")
	}
	time.Sleep(time.Second * 2)
	cache, err = doccache.New(dg, admin, client, cfg, nil)
	if err != nil {
		log.Fatal(err, "Failed creating DocCache")
	}
}

func TestReloadSchema(t *testing.T) {
	setUp("./config-no-special-config.yml")
	cache2, err := doccache.New(dg, admin, client, cfg, nil)
	if err != nil {
		log.Fatal(err, "Failed creating DocCache")
	}

	assertDoccacheConfig(t, cache2, cfg)
	assert.Equal(t, cache2.Cursor.GetValue("id").(string), doccache.CursorIdValue)
}

func TestDoccacheConfigIsProperlyConfiguredForNoElasticEndpoint(t *testing.T) {
	setUp("./config-no-elastic-endpoint.yml")
	assertDoccacheConfig(t, cache, cfg)
}
func TestNameFieldIndexesUpdateFromExistingSchema(t *testing.T) {
	setUp("./config-no-special-config.yml")

	t.Logf("Creating schema with documents with name fields only having regex index")

	dhoId := "2"
	dhoType := gql.NewSimplifiedType(
		"Dho",
		map[string]*gql.SimplifiedField{
			"details_rootNode_n": {
				Name:    "details_rootNode_n",
				Type:    "String",
				Indexes: gql.NewIndexes("regexp"),
			},
			"details_hvoiceSalaryPerPhase_a": {
				Name:    "details_hvoiceSalaryPerPhase_a",
				Type:    "String",
				Indexes: gql.NewIndexes("term"),
			},
			"details_timeShareX100_i": {
				Name:    "details_timeShareX100_i",
				Type:    "Int64",
				Indexes: gql.NewIndexes("int64"),
			},
			"details_strToInt_s": {
				Name:    "details_strToInt_s",
				Type:    "String",
				Indexes: gql.NewIndexes("regexp"),
			},
			"system_originalApprovedDate_t": {
				Name:    "system_originalApprovedDate_t",
				Type:    gql.GQLType_Time,
				Indexes: gql.NewIndexes("hour"),
			},
		},
		gql.DocumentSimplifiedInterface,
	)
	dhoInstance := gql.NewSimplifiedInstance(
		dhoType,
		map[string]interface{}{
			"docId":                          dhoId,
			"createdDate":                    "2020-11-12T18:27:47.000Z",
			"updatedDate":                    "2020-11-12T18:27:47.000Z",
			"creator":                        "dao.hypha",
			"contract":                       "contract1",
			"type":                           "Dho",
			"details_rootNode_n":             "dao.hypha2",
			"details_hvoiceSalaryPerPhase_a": "4133.04 HVOICE",
			"details_timeShareX100_i":        int64(60),
			"details_strToInt_s":             "60",
			"system_originalApprovedDate_t":  "2021-04-12T05:09:36.5Z",
		},
	)

	updateOp, err := cache.Schema.UpdateType(dhoType)
	assert.NilError(t, err)
	assert.Equal(t, updateOp, gql.SchemaUpdateOp_Created)
	// fmt.Println("Schema: ", schema.String())

	err = admin.UpdateSchema(cache.Schema)
	assert.NilError(t, err)
	numDocs := 500000
	fmt.Println("Inserting documents to test indexing speed: ", numDocs, time.Now())
	t.Logf("Inserting %v documents to test indexing speed", numDocs)
	for i := 0; i < numDocs; i++ {
		dhoInstance.Values["docId"] = fmt.Sprintf("%v", i)
		dhoInstance.SetValue("details_rootNode_n", fmt.Sprintf("dao.hypha%v", i))
		err = client.Mutate(dhoInstance.AddMutation(false))
		assert.NilError(t, err)
	}

	cache2, err := doccache.New(dg, admin, client, cfg, nil)
	if err != nil {
		log.Fatal(err, "Failed creating DocCache")
	}

	cache = cache2
	assertInstance(t, dhoInstance)

	fmt.Println("Updating document with name fields having both exact and regex indexes", time.Now())
	t.Logf("Updating document with name fields having both exact and regex indexes")
	dhoIdI, _ := strconv.ParseUint(dhoId, 10, 64)
	dhoDoc := &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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

	dhoType.Fields["details_rootNode_n"].Indexes = gql.NewIndexes("exact", "regexp")
	dhoInstance.SetValue("docId", dhoId)
	dhoInstance.SetValue("details_rootNode_n", "dao.hypha")
	cursor := "cursor1"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, dhoInstance)
	assertCursor(t, cursor)

	fmt.Println("Adding field to dho document to cause another schema update after adding index", time.Now())
	t.Logf("Adding field to dho document to cause another schema update after adding index")
	dhoId = "3"
	dhoIdI, _ = strconv.ParseUint(dhoId, 10, 64)
	dhoDoc = &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
					Label: "admin",
					Value: []interface{}{
						"name",
						"dao.admin",
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
	dhoType.SetField("details_admin_n", &gql.SimplifiedField{
		Name:    "details_admin_n",
		Type:    "String",
		Indexes: gql.NewIndexes("exact", "regexp"),
	})
	dhoInstance.SetValue("docId", dhoId)
	dhoInstance.SetValue("details_admin_n", "dao.admin")
	cursor = "cursor2"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, dhoInstance)
	assertCursor(t, cursor)

}

func TestOpCycle(t *testing.T) {
	setUp("./config-no-special-config.yml")
	assertDoccacheConfig(t, cache, cfg)
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing period 1 document")
	period1Id := "21"
	period1Hash := "h4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	period1IdI, _ := strconv.ParseUint(period1Id, 10, 64)
	periodDoc := getPeriodDoc(period1IdI, 1)
	expectedPeriodInstance := getPeriodInstance(period1IdI, 1)

	cursor := "cursor0"
	err := cache.StoreDocument(periodDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedPeriodInstance)
	assertCursor(t, cursor)

	t.Logf("Storing dho document")
	dhoId := "2"
	dhoIdI, _ := strconv.ParseUint(dhoId, 10, 64)
	dhoDoc := &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
	expectedDhoType := gql.NewSimplifiedType(
		"Dho",
		map[string]*gql.SimplifiedField{
			"details_rootNode_n": {
				Name:    "details_rootNode_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact", "regexp"),
			},
			"details_hvoiceSalaryPerPhase_a": {
				Name:    "details_hvoiceSalaryPerPhase_a",
				Type:    "String",
				Indexes: gql.NewIndexes("term"),
			},
			"details_timeShareX100_i": {
				Name:    "details_timeShareX100_i",
				Type:    "Int64",
				Indexes: gql.NewIndexes("int64"),
			},
			"details_strToInt_s": {
				Name:    "details_strToInt_s",
				Type:    "String",
				Indexes: gql.NewIndexes("regexp"),
			},
			"details_startPeriod_c": {
				Name:    "details_startPeriod_c",
				Type:    "String",
				Indexes: gql.NewIndexes("exact"),
			},
			"system_originalApprovedDate_t": {
				Name:    "system_originalApprovedDate_t",
				Type:    gql.GQLType_Time,
				Indexes: gql.NewIndexes("hour"),
			},
		},
		gql.DocumentSimplifiedInterface,
	)
	expectedDHOInstance := gql.NewSimplifiedInstance(
		expectedDhoType,
		map[string]interface{}{
			"docId":                          dhoId,
			"createdDate":                    "2020-11-12T18:27:47.000Z",
			"updatedDate":                    "2020-11-12T18:27:47.000Z",
			"creator":                        "dao.hypha",
			"contract":                       "contract1",
			"type":                           "Dho",
			"details_rootNode_n":             "dao.hypha",
			"details_hvoiceSalaryPerPhase_a": "4133.04 HVOICE",
			"details_timeShareX100_i":        int64(60),
			"details_strToInt_s":             "60",
			"details_startPeriod_c":          period1Hash,
			"system_originalApprovedDate_t":  "2021-04-12T05:09:36.5Z",
		},
	)
	cursor = "cursor1"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing member document")
	member1Id := "31"
	member1IdI, _ := strconv.ParseUint(member1Id, 10, 64)
	memberDoc := getMemberDoc(member1IdI, "member1")
	expectedMemberInstance := getMemberInstance(member1IdI, "member1")
	cursor = "cursor2_1"

	err = cache.StoreDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedMemberInstance)
	assertCursor(t, cursor)

	t.Logf("Storing another member document")
	member2Id := "32"
	member2IdI, _ := strconv.ParseUint(member2Id, 10, 64)
	memberDoc = getMemberDoc(member2IdI, "member2")
	expectedMemberInstance = getMemberInstance(member2IdI, "member2")
	cursor = "cursor2_2"

	err = cache.StoreDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedMemberInstance)
	assertCursor(t, cursor)

	t.Logf("Storing user document")
	user1Id := "41"
	user1IdI, _ := strconv.ParseUint(user1Id, 10, 64)
	userDoc := getUserDoc(user1IdI, "user1")
	expectedUserInstance := getUserInstance(user1IdI, "user1")
	cursor = "cursor3"

	err = cache.StoreDocument(userDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedUserInstance)
	assertCursor(t, cursor)

	t.Log("Adding member edge1")
	cursor = "cursor4_1"
	err = cache.MutateEdge(domain.NewChainEdge("member", dhoId, member1Id), false, cursor)
	assert.NilError(t, err)

	expectedDhoType.SetField("member", &gql.SimplifiedField{
		Name:    "member",
		Type:    "Member",
		IsArray: true,
		NonNull: false,
	})
	expectedMemberEdge := []map[string]interface{}{
		{"docId": member1Id},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Adding member edge2")
	cursor = "cursor4_2"
	err = cache.MutateEdge(domain.NewChainEdge("member", dhoId, member2Id), false, cursor)
	assert.NilError(t, err)

	expectedMemberEdge = []map[string]interface{}{
		{"docId": member1Id},
		{"docId": member2Id},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Adding user edge1, should cause edge type to change to doc")
	cursor = "cursor4_2"
	err = cache.MutateEdge(domain.NewChainEdge("member", dhoId, user1Id), false, cursor)
	assert.NilError(t, err)

	expectedDhoType.SetField("member", &gql.SimplifiedField{
		Name:    "member",
		Type:    "Document",
		IsArray: true,
		NonNull: false,
	})

	expectedMemberEdge = []map[string]interface{}{
		{"docId": member1Id},
		{"docId": member2Id},
		{"docId": user1Id},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing period 2 document")
	period2Id := "22"
	period2IdI, _ := strconv.ParseUint(period2Id, 10, 64)
	period2Hash := "i4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	periodDoc = getPeriodDoc(period2IdI, 2)
	expectedPeriodInstance = getPeriodInstance(period2IdI, 2)

	cursor = "cursorA"
	err = cache.StoreDocument(periodDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedPeriodInstance)
	assertCursor(t, cursor)

	t.Log("Update DHO document: Update values, add coreedge, remove core field")
	dhoDoc = &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T20:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
			Name:    "details_strToInt_i",
			Type:    "Int64",
			Indexes: gql.NewIndexes("int64"),
		},
	)
	expectedDhoType.SetField(
		"details_periodCount_i",
		&gql.SimplifiedField{
			Name:    "details_periodCount_i",
			Type:    "Int64",
			Indexes: gql.NewIndexes("int64"),
		},
	)
	expectedDhoType.SetField(
		"system_endPeriod_c",
		&gql.SimplifiedField{
			Name:    "system_endPeriod_c",
			Type:    "String",
			Indexes: gql.NewIndexes("exact"),
		},
	)
	expectedDHOInstance.SetValue("updatedDate", "2020-11-12T20:27:47.000Z")
	expectedDHOInstance.SetValue("details_periodCount_i", int64(50))
	expectedDHOInstance.SetValue("details_timeShareX100_i", nil)
	expectedDHOInstance.SetValue("details_strToInt_s", nil)
	expectedDHOInstance.SetValue("details_strToInt_i", int64(60))
	expectedDHOInstance.SetValue("system_originalApprovedDate_t", "2021-05-12T05:09:36.5Z")
	expectedDHOInstance.SetValue("details_hvoiceSalaryPerPhase_a", "4233.04 HVOICE")
	expectedDHOInstance.SetValue("system_endPeriod_c", period2Hash)

	cursor = "cursor6"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing period 3 document")
	period3Id := "23"
	period3IdI, _ := strconv.ParseUint(period3Id, 10, 64)
	period3Hash := "i3fc74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	periodDoc = getPeriodDoc(period3IdI, 3)
	expectedPeriodInstance = getPeriodInstance(period3IdI, 3)

	cursor = "cursor7_1"
	err = cache.StoreDocument(periodDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedPeriodInstance)
	assertCursor(t, cursor)

	t.Log("Update 2 dho, update core edge")
	dhoDoc = &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T21:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
						period3Hash,
					},
				},
			},
		},
	}

	expectedDHOInstance.SetValue("updatedDate", "2020-11-12T21:27:47.000Z")
	expectedDHOInstance.SetValue("system_endPeriod_c", period3Hash)

	cursor = "cursor6"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Deleting member 1 edge")
	cursor = "cursor7"
	err = cache.MutateEdge(domain.NewChainEdge("member", dhoId, member1Id), true, cursor)
	assert.NilError(t, err)

	expectedMemberEdge = []map[string]interface{}{
		{"docId": member2Id},
		{"docId": user1Id},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)

	t.Log("Update 3 DHO document: remove core edge")
	dhoDoc = &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T22:28:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
						period3Hash,
					},
				},
			},
		},
	}

	expectedDHOInstance.SetValue("updatedDate", "2020-11-12T22:28:47.000Z")
	expectedDHOInstance.SetValue("details_startPeriod_c", nil)

	cursor = "cursorB"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Log("Deleting user 1 edge")
	cursor = "cursor7_1"
	err = cache.MutateEdge(domain.NewChainEdge("member", dhoId, user1Id), true, cursor)
	assert.NilError(t, err)

	expectedMemberEdge = []map[string]interface{}{
		{"docId": member2Id},
	}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)

	t.Log("Deleting member2 edge")
	cursor = "cursor8"
	err = cache.MutateEdge(domain.NewChainEdge("member", dhoId, member2Id), true, cursor)
	assert.NilError(t, err)

	expectedMemberEdge = []map[string]interface{}{}
	expectedDHOInstance.SetValue("member", expectedMemberEdge)
	assertInstance(t, expectedDHOInstance)

	t.Logf("Deleting user1 document")
	userDoc = getUserDoc(user1IdI, "user1")
	cursor = "cursor8_1"

	err = cache.DeleteDocument(userDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, user1Id, "User")
	assertCursor(t, cursor)

	t.Logf("Deleting member1 document")
	memberDoc = getMemberDoc(member1IdI, "member1")
	cursor = "cursor9"

	err = cache.DeleteDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, member1Id, "Member")
	assertCursor(t, cursor)

	t.Logf("Deleting member2 document")
	memberDoc = getMemberDoc(member2IdI, "member2")
	cursor = "cursor10"

	err = cache.DeleteDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, member2Id, "Member")
	assertCursor(t, cursor)

	t.Logf("Deleting dho document")
	cursor = "cursor11"
	err = cache.DeleteDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, dhoId, "Dho")
	assertCursor(t, cursor)

}

func TestDocumentCreationDeduceType(t *testing.T) {
	setUp("./config-with-special-config.yml")
	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T19:27:47.000"
	chainDoc1Id := "1"
	chainDoc1IdI, _ := strconv.ParseUint(chainDoc1Id, 10, 64)
	chainDoc1 := &domain.ChainDocument{
		ID:          chainDoc1IdI,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
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

	expectedInstance := gql.NewSimplifiedInstance(
		gql.NewSimplifiedType(
			"VoteTally",
			map[string]*gql.SimplifiedField{
				"pass_votePower_a": {
					Name:    "pass_votePower_a",
					Type:    "String",
					Indexes: gql.NewIndexes("term"),
				},
				"fail_votePower_a": {
					Name:    "fail_votePower_a",
					Type:    "String",
					Indexes: gql.NewIndexes("term"),
				},
			},
			gql.DocumentSimplifiedInterface,
		),
		map[string]interface{}{
			"docId":            chainDoc1Id,
			"createdDate":      "2020-11-12T18:27:47.000Z",
			"updatedDate":      "2020-11-12T19:27:47.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "VoteTally",
			"pass_votePower_a": "0.00 HVOICE",
			"fail_votePower_a": "1.00 HVOICE",
		},
	)

	cursor := "cursor0"
	err := cache.StoreDocument(chainDoc1, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance)
	assertCursor(t, cursor)

	cursor = "cursor1"

	err = cache.DeleteDocument(chainDoc1, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, chainDoc1Id, "VoteTally")
	assertCursor(t, cursor)

}

func TestMissingCoreEdge(t *testing.T) {
	setUp("./config-no-special-config.yml")
	t.Log("Store assignment 1 with related core edge non existant")
	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T19:27:47.000"
	period1Id := "21"
	period1IdI, _ := strconv.ParseUint(period1Id, 10, 64)
	period1Hash := "a5ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"

	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1 := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
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

	expectedType := gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
			"details_startPeriod_c": {
				Name:    "details_startPeriod_c",
				Type:    "String",
				Indexes: gql.NewIndexes("exact"),
			},
		},
		gql.DocumentSimplifiedInterface,
	)

	expectedInstance := gql.NewSimplifiedInstance(
		expectedType,
		map[string]interface{}{
			"docId":                 assignment1Id,
			"createdDate":           "2020-11-12T18:27:47.000Z",
			"updatedDate":           "2020-11-12T19:27:47.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "Assignment",
			"details_startPeriod_c": period1Hash,
		},
	)

	cursor := "cursor0"
	err := cache.StoreDocument(assignment1, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance)
	assertCursor(t, cursor)

	t.Log("Store core edge")

	period1Doc := getPeriodDoc(period1IdI, 1)
	period1Instance := getPeriodInstance(period1IdI, 1)
	cursor = "cursor1"
	err = cache.StoreDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, period1Instance)
	assertCursor(t, cursor)

	t.Log("Store assignment 2 with related core edge")
	assignment2Id := "2"
	assignment2IdI, _ := strconv.ParseUint(assignment2Id, 10, 64)
	assignment2 := &domain.ChainDocument{
		ID:          assignment2IdI,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
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

	expectedInstance2 := gql.NewSimplifiedInstance(
		expectedType,
		map[string]interface{}{
			"docId":                 assignment2Id,
			"createdDate":           "2020-11-12T18:27:47.000Z",
			"updatedDate":           "2020-11-12T19:27:47.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "Assignment",
			"details_startPeriod_c": period1Hash,
		},
	)

	cursor = "cursor2"
	err = cache.StoreDocument(assignment2, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedInstance2)
	assertCursor(t, cursor)

	cursor = "cursor4"

	t.Log("Delete core edge document")
	err = cache.DeleteDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, period1Id, "Period")
	assertCursor(t, cursor)

	t.Log("Store core edge again")
	cursor = "cursor5"
	err = cache.StoreDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, period1Instance)
	assertCursor(t, cursor)

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
	assertInstanceNotExists(t, assignment1Id, "Assignment")
	assertCursor(t, cursor)

	cursor = "cursor8"
	err = cache.DeleteDocument(assignment2, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, assignment2Id, "Assignment")
	assertCursor(t, cursor)

	cursor = "cursor9"
	err = cache.DeleteDocument(period1Doc, cursor)
	assert.NilError(t, err)
	assertInstanceNotExists(t, period1Id, "Period")
	assertCursor(t, cursor)

}

func TestLogicalIds(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing dho1 document")
	dhoId := "2"
	dhoIdI, _ := strconv.ParseUint(dhoId, 10, 64)
	dhoDoc := &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
			},
		},
	}
	expectedDhoType := gql.NewSimplifiedType(
		"Dho",
		map[string]*gql.SimplifiedField{
			"details_rootNode_n": {
				Name:    "details_rootNode_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact", "regexp"),
				IsID:    true,
				NonNull: true,
			},
			"details_hvoiceSalaryPerPhase_a": {
				Name:    "details_hvoiceSalaryPerPhase_a",
				Type:    "String",
				Indexes: gql.NewIndexes("term"),
			},
		},
		gql.DocumentSimplifiedInterface,
	)
	expectedDHOInstance := gql.NewSimplifiedInstance(
		expectedDhoType,
		map[string]interface{}{
			"docId":                          dhoId,
			"createdDate":                    "2020-11-12T18:27:47.000Z",
			"updatedDate":                    "2020-11-12T19:27:47.000Z",
			"creator":                        "dao.hypha",
			"contract":                       "contract1",
			"type":                           "Dho",
			"details_rootNode_n":             "dao.hypha",
			"details_hvoiceSalaryPerPhase_a": "4133.04 HVOICE",
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing dho2 document")
	dhoId = "3"
	dhoIdI, _ = strconv.ParseUint(dhoId, 10, 64)
	dhoDoc = &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "root_node",
					Value: []interface{}{
						"name",
						"dao.beta",
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
						"4133.14 HVOICE",
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
			},
		},
	}
	expectedDHOInstance = gql.NewSimplifiedInstance(
		expectedDhoType,
		map[string]interface{}{
			"docId":                          dhoId,
			"createdDate":                    "2020-11-12T18:27:47.000Z",
			"updatedDate":                    "2020-11-12T19:27:47.000Z",
			"creator":                        "dao.hypha",
			"contract":                       "contract1",
			"type":                           "Dho",
			"details_rootNode_n":             "dao.beta",
			"details_hvoiceSalaryPerPhase_a": "4133.14 HVOICE",
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(dhoDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedDHOInstance)
	assertCursor(t, cursor)

	t.Logf("Storing member document")
	memberId := "31"
	memberIdI, _ := strconv.ParseUint(memberId, 10, 64)
	memberDoc := &domain.ChainDocument{
		ID:          memberIdI,
		CreatedDate: "2020-11-12T19:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "bob",
		Contract:    "contract1",
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
					Label: "member",
					Value: []interface{}{
						"name",
						"bob",
					},
				},
				{
					Label: "root_node",
					Value: []interface{}{
						"name",
						"dao.beta",
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
	expectedMemberType := gql.NewSimplifiedType(
		"Member",
		map[string]*gql.SimplifiedField{
			"details_rootNode_n": {
				Name:    "details_rootNode_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact", "regexp"),
			},
			"details_member_n": {
				Name:    "details_member_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact", "regexp"),
				IsID:    true,
				NonNull: true,
			},
		},
		gql.DocumentSimplifiedInterface,
	)
	expectedMemberInstance := gql.NewSimplifiedInstance(
		expectedMemberType,
		map[string]interface{}{
			"docId":              memberId,
			"createdDate":        "2020-11-12T19:27:47.000Z",
			"updatedDate":        "2020-11-12T19:27:47.000Z",
			"creator":            "bob",
			"contract":           "contract1",
			"type":               "Member",
			"details_rootNode_n": "dao.beta",
			"details_member_n":   "bob",
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(memberDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedMemberInstance)
	assertCursor(t, cursor)
}

func TestLogicalIdsShouldFailForNonExistantId(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing dho1 document")
	dhoId := "1"
	dhoIdI, _ := strconv.ParseUint(dhoId, 10, 64)
	dhoDoc := &domain.ChainDocument{
		ID:          dhoIdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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
					Label: "hvoice_salary_per_phase",
					Value: []interface{}{
						"asset",
						"4133.04 HVOICE",
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
			},
		},
	}
	cursor := "cursor1"
	err := cache.StoreDocument(dhoDoc, cursor)
	assert.ErrorContains(t, err, "does not have logical id field")

}

func TestCustomInterfaceInitialization(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Custom Interfaces and associated types should have been created during doccache initialization")

	currentSchema, err := admin.GetCurrentSchema()

	t.Logf("Checking Vote Type related to Votable interface")
	voteType := gql.NewSimplifiedType(
		"Vote",
		map[string]*gql.SimplifiedField{},
		gql.DocumentSimplifiedInterface,
	)
	assert.NilError(t, err)
	util.AssertType(t, voteType, currentSchema)

	t.Logf("Checking VoteTally Type related to Votable interface")
	voteTallyType := gql.NewSimplifiedType(
		"VoteTally",
		map[string]*gql.SimplifiedField{},
		gql.DocumentSimplifiedInterface,
	)
	assert.NilError(t, err)
	util.AssertType(t, voteTallyType, currentSchema)

	t.Logf("Checking Votable interface")
	votableInterface := gql.NewSimplifiedInterface(
		"Votable",
		map[string]*gql.SimplifiedField{
			"ballot_expiration_t": {
				Name:    "ballot_expiration_t",
				Type:    gql.GQLType_Time,
				Indexes: gql.NewIndexes("hour"),
			},
			"details_title_s": {
				Name:    "details_title_s",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("regexp"),
				IsID:    true,
				NonNull: true,
			},
			"details_description_s": {
				Name:    "details_description_s",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("regexp"),
			},
			"vote": {
				Name:    "vote",
				Type:    "Vote",
				IsArray: true,
			},
			"votetally": {
				Name:    "votetally",
				Type:    "VoteTally",
				IsArray: true,
			},
		},
		[]string{
			"ballot_expiration_t",
		},
		[]string{
			"Payout",
			"AssignBadge",
		},
	)
	util.AssertInterface(t, votableInterface, currentSchema)

	// t.Logf("Checking Profile Type related to User interface")
	// profileDataType := gql.NewSimplifiedType(
	// 	"ProfileData",
	// 	map[string]*gql.SimplifiedField{},
	// 	gql.DocumentSimplifiedInterface,
	// )
	// assert.NilError(t, err)
	// util.AssertType(t, profileDataType, currentSchema)

	t.Logf("Checking User interface")
	userInterface := gql.NewSimplifiedInterface(
		"User",
		map[string]*gql.SimplifiedField{
			"details_profile_c": {
				Name:    "details_profile_c",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("exact"),
			},
			"details_account_n": {
				Name:    "details_account_n",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("regexp"),
			},
		},
		[]string{
			"details_profile_c",
			"details_account_s",
		},
		nil,
	)
	util.AssertInterface(t, userInterface, currentSchema)

	t.Logf("Checking Extendable interface")
	extendableInterface := gql.NewSimplifiedInterface(
		"Extendable",
		map[string]*gql.SimplifiedField{
			"details_extensionName_s": {
				Name:    "details_extensionName_s",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("regexp"),
			},
			"extension": {
				Name:    "extension",
				Type:    "Document",
				IsArray: true,
			},
		},
		[]string{
			"details_extensionName_s",
		},
		nil,
	)
	util.AssertInterface(t, extendableInterface, currentSchema)

	t.Logf("Checking Taskable interface")
	taskableInterface := gql.NewSimplifiedInterface(
		"Taskable",
		map[string]*gql.SimplifiedField{
			"details_task_s": {
				Name:    "details_task_s",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("regexp"),
			},
			"user": {
				Name:    "user",
				Type:    "User",
				IsArray: true,
			},
		},
		[]string{},
		[]string{
			"AdminTask",
		},
	)
	util.AssertInterface(t, taskableInterface, currentSchema)

	t.Logf("Checking Editable interface")
	editableInterface := gql.NewSimplifiedInterface(
		"Editable",
		map[string]*gql.SimplifiedField{
			"details_version_s": {
				Name:    "details_version_s",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("regexp"),
			},
		},
		[]string{},
		[]string{
			"AdminTask",
			"Payout",
		},
	)
	util.AssertInterface(t, editableInterface, currentSchema)

}

func TestCustomInterfaces(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assignment proposal 1 document, has signature fields so it should implement Votable interface")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "expiration",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:27:47.000",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment1Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                 assignment1Id,
			"createdDate":           "2020-11-12T18:27:47.000Z",
			"updatedDate":           "2020-11-12T19:27:47.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "AssigProp",
			"ballot_expiration_t":   "2020-11-15T18:27:47.000Z",
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(assignment1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment proposal 2 document, does not have signature fields but because the assignment type already has the interface it should implement it")
	assignment2Id := "2"
	assignment2IdI, _ := strconv.ParseUint(assignment2Id, 10, 64)
	assignment2Doc := &domain.ChainDocument{
		ID:          assignment2IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "started_at",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:28:47.000",
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 2",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType = &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_startedAt_t": {
					Name:    "details_startedAt_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment2Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                 assignment2Id,
			"createdDate":           "2020-11-12T18:27:48.000Z",
			"updatedDate":           "2020-11-12T19:27:48.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "AssigProp",
			"ballot_expiration_t":   nil,
			"details_startedAt_t":   "2020-11-15T18:28:47.000Z",
			"details_title_s":       "Assignment 2",
			"details_description_s": nil,
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(assignment2Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment2Instance)
	assertCursor(t, cursor)

	t.Logf("Storing profile data document to be used as core edge")
	profileId := "21"
	profileHash := "c4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	profileIdI, _ := strconv.ParseUint(profileId, 10, 64)
	profileDoc := &domain.ChainDocument{
		ID:          profileIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "name",
					Value: []interface{}{
						"string",
						"User 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"profile.data",
					},
				},
			},
		},
	}
	expectedProfileType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "ProfileData",
			Fields: map[string]*gql.SimplifiedField{
				"details_name_s": {
					Name:    "details_name_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedProfileType.SetFields(gql.DocumentFieldArgs)
	expectedProfileInstance := gql.NewSimplifiedInstance(
		expectedProfileType,
		map[string]interface{}{
			"docId":          profileId,
			"createdDate":    "2020-11-12T18:27:48.000Z",
			"updatedDate":    "2020-11-12T19:27:48.000Z",
			"creator":        "dao.hypha",
			"contract":       "contract1",
			"type":           "ProfileData",
			"details_name_s": "User 1",
		},
	)
	cursor = "cursor3"
	err = cache.StoreDocument(profileDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedProfileInstance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment proposal 3 document, has signature fields for User Interface, but as its an old type it should NOT be added")
	assignment3Id := "3"
	assignment3IdI, _ := strconv.ParseUint(assignment3Id, 10, 64)
	assignment3Doc := &domain.ChainDocument{
		ID:          assignment3IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "profile",
					Value: []interface{}{
						"checksum256",
						profileHash,
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 3",
					},
				},
				{
					Label: "account",
					Value: []interface{}{
						"name",
						"user1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType = &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_startedAt_t": {
					Name:    "details_startedAt_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
				"details_profile_c": {
					Name:    "details_profile_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},
				"details_account_n": {
					Name:    "details_account_n",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact", "regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment3Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                 assignment3Id,
			"createdDate":           "2020-11-12T18:27:48.000Z",
			"updatedDate":           "2020-11-12T19:27:48.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "AssigProp",
			"ballot_expiration_t":   nil,
			"details_startedAt_t":   nil,
			"details_title_s":       "Assignment 3",
			"details_description_s": nil,
			"details_profile_c":     profileHash,
			"details_account_n":     "user1",
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor = "cursor4"
	err = cache.StoreDocument(assignment3Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment3Instance)
	assertCursor(t, cursor)

	t.Logf("Storing vote document to be used as edge that is part of the interface")
	voteId := "41"
	voteIdI, _ := strconv.ParseUint(voteId, 10, 64)
	voteDoc := &domain.ChainDocument{
		ID:          voteIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "result",
					Value: []interface{}{
						"string",
						"For",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"vote",
					},
				},
			},
		},
	}
	expectedVoteType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Vote",
			Fields: map[string]*gql.SimplifiedField{
				"details_result_s": {
					Name:    "details_result_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedVoteType.SetFields(gql.DocumentFieldArgs)
	expectedVoteInstance := gql.NewSimplifiedInstance(
		expectedVoteType,
		map[string]interface{}{
			"docId":            voteId,
			"createdDate":      "2020-11-12T18:27:48.000Z",
			"updatedDate":      "2020-11-12T19:27:48.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "Vote",
			"details_result_s": "For",
		},
	)
	cursor = "cursor6"
	err = cache.StoreDocument(voteDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedVoteInstance)
	assertCursor(t, cursor)

	t.Log("Adding vote edge")
	cursor = "cursor7"
	err = cache.MutateEdge(domain.NewChainEdge("vote", assignment3Id, voteId), false, cursor)
	assert.NilError(t, err)

	expectedVoteEdge := []map[string]interface{}{
		{"docId": voteId},
	}
	expectedAssignment3Instance.SetValue("vote", expectedVoteEdge)
	assertInstance(t, expectedAssignment3Instance)
	assertCursor(t, cursor)

}

func TestCustomInterfacesAddByType(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assign badge proposal document, is of votable type")
	assignBadge1Id := "1"
	assignBadge1IdI, _ := strconv.ParseUint(assignBadge1Id, 10, 64)
	assignBadge1Doc := &domain.ChainDocument{
		ID:          assignBadge1IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						11,
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assignbadge",
					},
				},
			},
		},
	}
	expectedAssignBadgeType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Assignbadge",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"ballot_votes_i": {
					Name:    "ballot_votes_i",
					Type:    gql.GQLType_Int64,
					Indexes: gql.NewIndexes("int64"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable"},
	}
	expectedAssignBadgeType.SetFields(gql.DocumentFieldArgs)
	expectedAssignBadge1Instance := gql.NewSimplifiedInstance(
		expectedAssignBadgeType,
		map[string]interface{}{
			"docId":                 assignBadge1Id,
			"createdDate":           "2020-11-12T18:27:48.000Z",
			"updatedDate":           "2020-11-12T19:27:48.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "Assignbadge",
			"ballot_expiration_t":   nil,
			"ballot_votes_i":        11,
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(assignBadge1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignBadge1Instance)
	assertCursor(t, cursor)

}

func TestCustomInterfacesAddMultipleByType(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assign badge proposal document, is of votable type")
	payout1Id := "1"
	payout1IdI, _ := strconv.ParseUint(payout1Id, 10, 64)
	payout1Doc := &domain.ChainDocument{
		ID:          payout1IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						11,
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"payout",
					},
				},
			},
		},
	}
	expectedPayoutType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Payout",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"ballot_votes_i": {
					Name:    "ballot_votes_i",
					Type:    gql.GQLType_Int64,
					Indexes: gql.NewIndexes("int64"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"details_version_s": {
					Name:    "details_version_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable", "Editable"},
	}
	expectedPayoutType.SetFields(gql.DocumentFieldArgs)
	expectedPayout1Instance := gql.NewSimplifiedInstance(
		expectedPayoutType,
		map[string]interface{}{
			"docId":                 payout1Id,
			"createdDate":           "2020-11-12T18:27:48.000Z",
			"updatedDate":           "2020-11-12T19:27:48.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "Payout",
			"ballot_expiration_t":   nil,
			"ballot_votes_i":        11,
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"details_version_s":     nil,
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(payout1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedPayout1Instance)
	assertCursor(t, cursor)

}

func TestCustomInterfacesAddSignatureAndTypeBased(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing profile data document to be used as core edge")
	profileId := "31"
	profileIdI, _ := strconv.ParseUint(profileId, 10, 64)
	profileHash := "c4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	profileDoc := &domain.ChainDocument{
		ID:          profileIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "name",
					Value: []interface{}{
						"string",
						"User 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"profile.data",
					},
				},
			},
		},
	}
	expectedProfileType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "ProfileData",
			Fields: map[string]*gql.SimplifiedField{
				"details_name_s": {
					Name:    "details_name_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedProfileType.SetFields(gql.DocumentFieldArgs)
	expectedProfileInstance := gql.NewSimplifiedInstance(
		expectedProfileType,
		map[string]interface{}{
			"docId":          profileId,
			"createdDate":    "2020-11-12T18:27:48.000Z",
			"updatedDate":    "2020-11-12T19:27:48.000Z",
			"creator":        "dao.hypha",
			"contract":       "contract1",
			"type":           "ProfileData",
			"details_name_s": "User 1",
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(profileDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedProfileInstance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment badge document, has type for Votable and signature fields for User interfaces, both should be added")
	assignbadge1Id := "1"
	assignbadge1IdI, _ := strconv.ParseUint(assignbadge1Id, 10, 64)
	assignbadge1Doc := &domain.ChainDocument{
		ID:          assignbadge1IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						11,
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "profile",
					Value: []interface{}{
						"checksum256",
						profileHash,
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "account",
					Value: []interface{}{
						"name",
						"user2",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assignbadge",
					},
				},
			},
		},
	}
	expectedAssignbadgeType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Assignbadge",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"ballot_votes_i": {
					Name:    "ballot_votes_i",
					Type:    gql.GQLType_Int64,
					Indexes: gql.NewIndexes("int64"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
				"details_profile_c": {
					Name:    "details_profile_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},

				"details_account_n": {
					Name:    "details_account_n",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact", "regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable", "User"},
	}
	expectedAssignbadgeType.SetFields(gql.DocumentFieldArgs)
	expectedAssignbadge1Instance := gql.NewSimplifiedInstance(
		expectedAssignbadgeType,
		map[string]interface{}{
			"docId":                 assignbadge1Id,
			"createdDate":           "2020-11-12T18:27:48.000Z",
			"updatedDate":           "2020-11-12T19:27:48.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "Assignbadge",
			"ballot_expiration_t":   nil,
			"ballot_votes_i":        11,
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"details_profile_c":     profileHash,
			"details_account_n":     "user2",
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(assignbadge1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignbadge1Instance)
	assertCursor(t, cursor)

}

func TestCustomInterfacesAddMultipleAtTheSameTime(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing profile data document to be used as core edge")
	profileId := "31"
	profileIdI, _ := strconv.ParseUint(profileId, 10, 64)
	profileHash := "c4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	profileDoc := &domain.ChainDocument{
		ID:          profileIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "name",
					Value: []interface{}{
						"string",
						"User 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"profile.data",
					},
				},
			},
		},
	}
	expectedProfileType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "ProfileData",
			Fields: map[string]*gql.SimplifiedField{
				"details_name_s": {
					Name:    "details_name_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedProfileType.SetFields(gql.DocumentFieldArgs)
	expectedProfileInstance := gql.NewSimplifiedInstance(
		expectedProfileType,
		map[string]interface{}{
			"docId":          profileId,
			"createdDate":    "2020-11-12T18:27:48.000Z",
			"updatedDate":    "2020-11-12T19:27:48.000Z",
			"creator":        "dao.hypha",
			"contract":       "contract1",
			"type":           "ProfileData",
			"details_name_s": "User 1",
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(profileDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedProfileInstance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment proposal document, has signature fields for Votable and User interfaces, both should be added")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "expiration",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:27:47.000",
					},
				},
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						11,
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "profile",
					Value: []interface{}{
						"checksum256",
						profileHash,
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "account",
					Value: []interface{}{
						"name",
						"user2",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"ballot_votes_i": {
					Name:    "ballot_votes_i",
					Type:    gql.GQLType_Int64,
					Indexes: gql.NewIndexes("int64"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
				"details_profile_c": {
					Name:    "details_profile_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},
				"details_account_n": {
					Name:    "details_account_n",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact", "regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable", "User"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment1Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                 assignment1Id,
			"createdDate":           "2020-11-12T18:27:48.000Z",
			"updatedDate":           "2020-11-12T19:27:48.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "AssigProp",
			"ballot_expiration_t":   "2020-11-15T18:27:47.000Z",
			"ballot_votes_i":        11,
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"details_profile_c":     profileHash,
			"details_account_n":     "user2",
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(assignment1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)

}
func TestCustomInterfacesWithCoreEdge(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing profile data document to be used as core edge")
	profileId := "31"
	profileIdI, _ := strconv.ParseUint(profileId, 10, 64)
	profileHash := "a4ec74355830056924c83f20ffb1a22ad0c5145a96daddf6301897a092de951e"
	profileDoc := &domain.ChainDocument{
		ID:          profileIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "name",
					Value: []interface{}{
						"string",
						"User 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"profile.data",
					},
				},
			},
		},
	}
	expectedProfileType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "ProfileData",
			Fields: map[string]*gql.SimplifiedField{
				"details_name_s": {
					Name:    "details_name_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedProfileType.SetFields(gql.DocumentFieldArgs)
	expectedProfileInstance := gql.NewSimplifiedInstance(
		expectedProfileType,
		map[string]interface{}{
			"docId":          profileId,
			"createdDate":    "2020-11-12T18:27:48.000Z",
			"updatedDate":    "2020-11-12T19:27:48.000Z",
			"creator":        "dao.hypha",
			"contract":       "contract1",
			"type":           "ProfileData",
			"details_name_s": "User 1",
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(profileDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedProfileInstance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment proposal 1 document, has signature fields for User Interface, it should be added")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "profile",
					Value: []interface{}{
						"checksum256",
						profileHash,
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 3",
					},
				},
				{
					Label: "account",
					Value: []interface{}{
						"name",
						"user1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"details_profile_c": {
					Name:    "details_profile_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},
				"details_account_n": {
					Name:    "details_account_n",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact", "regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "User"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment1Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":             assignment1Id,
			"createdDate":       "2020-11-12T18:27:48.000Z",
			"updatedDate":       "2020-11-12T19:27:48.000Z",
			"creator":           "dao.hypha",
			"contract":          "contract1",
			"type":              "AssigProp",
			"details_title_s":   "Assignment 3",
			"details_profile_c": profileHash,
			"details_account_n": "user1",
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(assignment1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)
}

func TestCustomInterfacesEdgeIsGeneralizedToDocument(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assignment proposal 1 document")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T19:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "extension_name",
					Value: []interface{}{
						"string",
						"Vote extension 1",
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 0",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"details_extensionName_s": {
					Name:    "details_extensionName_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"extension": {
					Name:    "extension",
					Type:    "Document",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Extendable"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment1Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                   assignment1Id,
			"createdDate":             "2020-11-12T19:27:47.000Z",
			"updatedDate":             "2020-11-12T19:27:47.000Z",
			"creator":                 "dao.hypha",
			"contract":                "contract1",
			"type":                    "AssigProp",
			"details_title_s":         "Assignment 0",
			"details_extensionName_s": "Vote extension 1",
			"extension":               make([]map[string]interface{}, 0),
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(assignment1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)

	t.Logf("Storing Vote document to be used as edge which type should be upgraded to document")
	voteId := "21"
	voteIdI, _ := strconv.ParseUint(voteId, 10, 64)
	voteDoc := &domain.ChainDocument{
		ID:          voteIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "result",
					Value: []interface{}{
						"string",
						"For",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"vote",
					},
				},
			},
		},
	}
	expectedVoteType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Vote",
			Fields: map[string]*gql.SimplifiedField{
				"details_result_s": {
					Name:    "details_result_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedVoteType.SetFields(gql.DocumentFieldArgs)
	expectedVoteInstance := gql.NewSimplifiedInstance(
		expectedVoteType,
		map[string]interface{}{
			"docId":            voteId,
			"createdDate":      "2020-11-12T18:27:48.000Z",
			"updatedDate":      "2020-11-12T19:27:48.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "Vote",
			"details_result_s": "For",
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(voteDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedVoteInstance)
	assertCursor(t, cursor)

	t.Log("Adding vote edge")
	cursor = "cursor7"
	err = cache.MutateEdge(domain.NewChainEdge("extension", assignment1Id, voteId), false, cursor)
	assert.NilError(t, err)

	expectedAssignmentType.SetField("extension", &gql.SimplifiedField{
		Name:    "extension",
		Type:    "Document",
		IsArray: true,
		NonNull: false,
	})

	expectedVoteEdge := []map[string]interface{}{
		{"docId": voteId},
	}
	expectedAssignment1Instance.SetValue("extension", expectedVoteEdge)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment proposal 2 document")
	assignment2Id := "2"
	assignment2IdI, _ := strconv.ParseUint(assignment2Id, 10, 64)
	assignment2Doc := &domain.ChainDocument{
		ID:          assignment2IdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "extension_name",
					Value: []interface{}{
						"string",
						"Vote extension",
					},
				},
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	cursor = "cursor2"
	err = cache.StoreDocument(assignment2Doc, cursor)
	assert.NilError(t, err)

	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment2Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                   assignment2Id,
			"createdDate":             "2020-11-12T18:27:47.000Z",
			"updatedDate":             "2020-11-12T19:27:47.000Z",
			"creator":                 "dao.hypha",
			"contract":                "contract1",
			"type":                    "AssigProp",
			"details_title_s":         "Assignment 1",
			"details_extensionName_s": "Vote extension",
			"extension":               make([]map[string]interface{}, 0),
		},
	)
	cursor = "cursor3"
	err = cache.StoreDocument(assignment2Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment2Instance)
	assertCursor(t, cursor)

	t.Logf("Checking assignment 1 instance is still valid")
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)
}

func TestCustomInterfacesShouldFailForTypeWithoutIDField(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assignment proposal 1 document without interface ID field")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T19:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						10,
					},
				},
				{
					Label: "expiration",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:27:47.000",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
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
						"assig.prop",
					},
				},
			},
		},
	}
	cursor := "cursor1"
	err := cache.StoreDocument(assignment1Doc, cursor)
	assert.ErrorContains(t, err, "can't add non null field")

}

func TestCustomInterfacesShouldFailForTypeThatImplementsInterfaceNotHavingIDField(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assignment proposal 1 document, has signature fields so it should implement Votable interface")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "expiration",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:27:47.000",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment1Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                 assignment1Id,
			"createdDate":           "2020-11-12T18:27:47.000Z",
			"updatedDate":           "2020-11-12T19:27:47.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "AssigProp",
			"ballot_expiration_t":   "2020-11-15T18:27:47.000Z",
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(assignment1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)

	t.Logf("Storing assignment proposal 2 document, does not have id field of implementing interface")
	assignment2Id := "2"
	assignment2IdI, _ := strconv.ParseUint(assignment2Id, 10, 64)
	assignment2Doc := &domain.ChainDocument{
		ID:          assignment2IdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "started_at",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:28:47.000",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}

	cursor = "cursor2"
	err = cache.StoreDocument(assignment2Doc, cursor)
	assert.ErrorContains(t, err, "can't add non null field")

}

func TestCustomInterfacesShouldFailForAddingInvalidTypeEdge(t *testing.T) {
	setUp("./config-with-special-config.yml")
	assert.Equal(t, cache.Cursor.GetValue("id").(string), doccache.CursorIdValue)

	t.Logf("Storing assignment proposal 1 document")
	assignment1Id := "1"
	assignment1IdI, _ := strconv.ParseUint(assignment1Id, 10, 64)
	assignment1Doc := &domain.ChainDocument{
		ID:          assignment1IdI,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "expiration",
					Value: []interface{}{
						"time_point",
						"2020-11-15T18:27:47.000",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"ballot",
					},
				},
			},
			{
				{
					Label: "title",
					Value: []interface{}{
						"string",
						"Assignment 1",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"assig.prop",
					},
				},
			},
		},
	}
	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "AssigProp",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:    "ballot_expiration_t",
					Type:    gql.GQLType_Time,
					Indexes: gql.NewIndexes("hour"),
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:    "details_description_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
				"vote": {
					Name:    "vote",
					Type:    "Vote",
					IsArray: true,
				},
				"votetally": {
					Name:    "votetally",
					Type:    "VoteTally",
					IsArray: true,
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document", "Votable"},
	}
	expectedAssignmentType.SetFields(gql.DocumentFieldArgs)
	expectedAssignment1Instance := gql.NewSimplifiedInstance(
		expectedAssignmentType,
		map[string]interface{}{
			"docId":                 assignment1Id,
			"createdDate":           "2020-11-12T18:27:47.000Z",
			"updatedDate":           "2020-11-12T19:27:47.000Z",
			"creator":               "dao.hypha",
			"contract":              "contract1",
			"type":                  "AssigProp",
			"ballot_expiration_t":   "2020-11-15T18:27:47.000Z",
			"details_title_s":       "Assignment 1",
			"details_description_s": nil,
			"vote":                  make([]map[string]interface{}, 0),
			"votetally":             make([]map[string]interface{}, 0),
		},
	)
	cursor := "cursor1"
	err := cache.StoreDocument(assignment1Doc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedAssignment1Instance)
	assertCursor(t, cursor)

	t.Logf("Storing VoteOld document to be used as edge that is incompatible with interface")
	voteId := "21"
	voteIdI, _ := strconv.ParseUint(voteId, 10, 64)
	voteDoc := &domain.ChainDocument{
		ID:          voteIdI,
		CreatedDate: "2020-11-12T18:27:48.000",
		UpdatedDate: "2020-11-12T19:27:48.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "result",
					Value: []interface{}{
						"string",
						"For",
					},
				},
				{
					Label: "content_group_label",
					Value: []interface{}{
						"string",
						"details",
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
						"vote.old",
					},
				},
			},
		},
	}
	expectedVoteType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "VoteOld",
			Fields: map[string]*gql.SimplifiedField{
				"details_result_s": {
					Name:    "details_result_s",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("regexp"),
				},
			},
			WithSubscription: true,
		},
		Interfaces: []string{"Document"},
	}
	expectedVoteType.SetFields(gql.DocumentFieldArgs)
	expectedVoteInstance := gql.NewSimplifiedInstance(
		expectedVoteType,
		map[string]interface{}{
			"docId":            voteId,
			"createdDate":      "2020-11-12T18:27:48.000Z",
			"updatedDate":      "2020-11-12T19:27:48.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "VoteOld",
			"details_result_s": "For",
		},
	)
	cursor = "cursor2"
	err = cache.StoreDocument(voteDoc, cursor)
	assert.NilError(t, err)
	assertInstance(t, expectedVoteInstance)
	assertCursor(t, cursor)

	t.Log("Adding vote edge")
	cursor = "cursor7"
	err = cache.MutateEdge(domain.NewChainEdge("vote", assignment1Id, voteId), false, cursor)
	assert.ErrorContains(t, err, "For type AssigProp to implement interface Votable the field vote must have type")

}

func assertCursor(t *testing.T, cursor string) {
	expected := gql.NewCursorInstance(doccache.CursorIdValue, cursor)
	actual, err := cache.GetCursorInstance(doccache.CursorIdValue, gql.CursorSimplifiedType, nil)
	assert.NilError(t, err)
	tutil.AssertSimplifiedInstance(t, actual, expected)
}

func assertDoccacheConfig(t *testing.T, cache *doccache.Doccache, cfg *config.Config) {
	expected := gql.NewSimplifiedInstance(
		gql.DoccacheConfigSimplifiedType,
		map[string]interface{}{
			"id":              "dc1",
			"contract":        cfg.ContractName,
			"eosEndpoint":     cfg.EosEndpoint,
			"documentsTable":  cfg.DocTableName,
			"edgesTable":      cfg.EdgeTableName,
			"elasticEndpoint": cfg.ElasticEndpoint,
			"elasticApiKey":   cfg.ElasticApiKey,
		},
	)
	actual, err := cache.GetDoccacheConfigInstance()
	assert.NilError(t, err)
	tutil.AssertSimplifiedInstance(t, actual, expected)
}

func assertInstance(t *testing.T, expected *gql.SimplifiedInstance) {
	actualType, err := cache.Schema.GetSimplifiedType(expected.SimplifiedType.Name)
	assert.NilError(t, err)
	actual, err := cache.GetDocumentInstance(expected.GetValue("docId"), actualType, nil)
	assert.NilError(t, err)
	// fmt.Println("Expected: ", expected)
	// fmt.Println("Actual: ", actual)
	tutil.AssertSimplifiedInstance(t, actual, expected)
}

func assertInstanceNotExists(t *testing.T, docId, typeName string) {
	actualType, err := cache.Schema.GetSimplifiedType(typeName)
	assert.NilError(t, err)
	actual, err := cache.GetDocumentInstance(docId, actualType, nil)
	assert.NilError(t, err)
	assert.Assert(t, actual == nil)
}

func getMemberDoc(docIdI uint64, account string) *domain.ChainDocument {
	return &domain.ChainDocument{
		ID:          docIdI,
		CreatedDate: "2020-11-12T19:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     account,
		Contract:    "contract1",
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

func getUserDoc(docIdI uint64, account string) *domain.ChainDocument {
	return &domain.ChainDocument{
		ID:          docIdI,
		CreatedDate: "2020-11-12T19:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     account,
		Contract:    "contract1",
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
						"user",
					},
				},
			},
		},
	}
}

func getUserInstance(docIdI uint64, account string) *gql.SimplifiedInstance {
	return gql.NewSimplifiedInstance(
		userType,
		map[string]interface{}{
			"docId":             strconv.FormatUint(docIdI, 10),
			"createdDate":       "2020-11-12T19:27:47.000Z",
			"updatedDate":       "2020-11-12T19:27:47.000Z",
			"creator":           account,
			"contract":          "contract1",
			"type":              "User",
			"details_account_n": account,
		},
	)
}

func getMemberInstance(docIdI uint64, account string) *gql.SimplifiedInstance {
	return gql.NewSimplifiedInstance(
		memberType,
		map[string]interface{}{
			"docId":             strconv.FormatUint(docIdI, 10),
			"createdDate":       "2020-11-12T19:27:47.000Z",
			"updatedDate":       "2020-11-12T19:27:47.000Z",
			"creator":           account,
			"contract":          "contract1",
			"type":              "Member",
			"details_account_n": account,
		},
	)
}

func getPeriodDoc(id uint64, number int64) *domain.ChainDocument {
	return &domain.ChainDocument{
		ID:          id,
		CreatedDate: "2020-11-12T18:27:47.000",
		UpdatedDate: "2020-11-12T19:27:47.000",
		Creator:     "dao.hypha",
		Contract:    "contract1",
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

func getPeriodInstance(docId uint64, number int64) *gql.SimplifiedInstance {
	return gql.NewSimplifiedInstance(
		periodType,
		map[string]interface{}{
			"docId":            strconv.FormatUint(docId, 10),
			"createdDate":      "2020-11-12T18:27:47.000Z",
			"updatedDate":      "2020-11-12T19:27:47.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "Period",
			"details_number_i": number,
		},
	)
}
