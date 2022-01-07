package domain_test

import (
	"encoding/json"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

func TestToParsedDoc(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
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
					Label: "role",
					Value: []interface{}{
						"checksum256",
						"b7cf9e60a6c33e79b32c2eeb4575857f3f2c4166e737c6b3863da62a2cfcf1cf",
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
				{
					Label: "period",
					Value: []interface{}{
						"checksum256",
						"f7cf9e60a6c33e79b32c2eeb4575857f3f2c4166e737c6b3863da62a2cfcf1cf",
					},
				},
			},
		},
	}
	parsedDoc, err := chainDoc1.ToParsedDoc(make(map[string][]string))
	assert.NilError(t, err)

	expectedSimplifiedInstance := gql.NewSimplifiedInstance(
		gql.NewSimplifiedType(
			"Dho",
			map[string]*gql.SimplifiedField{
				"details_rootNode_n": {
					Name:  "details_rootNode_n",
					Type:  "String",
					Index: "exact",
				},
				"details_role_c": {
					Name:  "details_role_c",
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
				"system_originalApprovedDate_t": {
					Name:  "system_originalApprovedDate_t",
					Type:  gql.GQLType_Time,
					Index: "hour",
				},
				"system_period_c": {
					Name:  "system_period_c",
					Type:  "String",
					Index: "exact",
				},
			},
			gql.DocumentSimplifiedInterface,
		),
		map[string]interface{}{
			"docId":                          "0",
			"docId_i":                        0,
			"createdDate":                    "2020-11-12T18:27:47.000Z",
			"updatedDate":                    "2020-11-12T20:37:47.000Z",
			"creator":                        "dao.hypha",
			"contract":                       "contract1",
			"type":                           "Dho",
			"details_rootNode_n":             "dao.hypha",
			"details_role_c":                 "b7cf9e60a6c33e79b32c2eeb4575857f3f2c4166e737c6b3863da62a2cfcf1cf",
			"details_hvoiceSalaryPerPhase_a": "4133.04 HVOICE",
			"details_timeShareX100_i":        int64(60),
			"system_originalApprovedDate_t":  "2021-04-12T05:09:36.5Z",
			"system_period_c":                "f7cf9e60a6c33e79b32c2eeb4575857f3f2c4166e737c6b3863da62a2cfcf1cf",
		},
	)

	expectedParsedDoc := &domain.ParsedDoc{
		Instance:       expectedSimplifiedInstance,
		ChecksumFields: []string{"details_role_c", "system_period_c"},
	}

	assertParsedDoc(t, parsedDoc, expectedParsedDoc)

	// certificationDate := "2020-11-12T20:27:47.000"
	// chainDoc1.Certificates = []*domain.ChainCertificate{
	// 	{
	// 		Certifier:         "sebastian",
	// 		Notes:             "Sebastian's Notes",
	// 		CertificationDate: certificationDate,
	// 	},
	// }

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

}

func TestToParsedDocDeduceType(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
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
						"0.00 HVOICE",
					},
				},
			},
		},
	}
	typeMappings := map[string][]string{
		"VoteTally": {
			"pass_votePower",
			"fail_votePower",
		},
	}
	parsedDoc, err := chainDoc1.ToParsedDoc(typeMappings)
	assert.NilError(t, err)

	expectedSimplifiedInstance := gql.NewSimplifiedInstance(
		gql.NewSimplifiedType(
			"VoteTally",
			map[string]*gql.SimplifiedField{
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
			gql.DocumentSimplifiedInterface,
		),
		map[string]interface{}{
			"docId":            "0",
			"docId_i":          0,
			"createdDate":      "2020-11-12T18:27:47.000Z",
			"updatedDate":      "2020-11-12T20:37:47.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "VoteTally",
			"pass_votePower_a": "0.00 HVOICE",
			"fail_votePower_a": "0.00 HVOICE",
		},
	)

	expectedParsedDoc := &domain.ParsedDoc{
		Instance:       expectedSimplifiedInstance,
		ChecksumFields: []string{},
	}

	assertParsedDoc(t, parsedDoc, expectedParsedDoc)
}

func TestToParsedDocDeduceTypeFailsForMissingFields(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
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
						"0.00 HVOICE",
					},
				},
			},
		},
	}
	typeMappings := map[string][]string{
		"VoteTally": {
			"pass_votePower",
			"fail_votePower",
		},
	}
	_, err := chainDoc1.ToParsedDoc(typeMappings)
	assert.ErrorContains(t, err, "document with ID: 0 does not have a type, and couldn't deduce from typeMappings")

}

func TestToParsedDocDotNamedType(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
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
						"0.00 HVOICE",
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
						"vote.tally",
					},
				},
			},
		},
	}
	typeMappings := map[string][]string{
		"VoteTally": {
			"pass_votePower",
			"fail_votePower",
		},
	}
	parsedDoc, err := chainDoc1.ToParsedDoc(typeMappings)
	assert.NilError(t, err)

	expectedSimplifiedInstance := gql.NewSimplifiedInstance(
		gql.NewSimplifiedType(
			"VoteTally",
			map[string]*gql.SimplifiedField{
				"fail_votePower_a": {
					Name:  "fail_votePower_a",
					Type:  "String",
					Index: "term",
				},
			},
			gql.DocumentSimplifiedInterface,
		),
		map[string]interface{}{
			"docId":            "0",
			"docId_i":          0,
			"createdDate":      "2020-11-12T18:27:47.000Z",
			"updatedDate":      "2020-11-12T20:37:47.000Z",
			"creator":          "dao.hypha",
			"contract":         "contract1",
			"type":             "VoteTally",
			"fail_votePower_a": "0.00 HVOICE",
		},
	)

	expectedParsedDoc := &domain.ParsedDoc{
		Instance:       expectedSimplifiedInstance,
		ChecksumFields: []string{},
	}

	assertParsedDoc(t, parsedDoc, expectedParsedDoc)
}

func TestToParsedDocShouldFailForNoContentGroupLabel(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
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
					Label: "role",
					Value: []interface{}{
						"checksum256",
						"b7cf9e60a6c33e79b32c2eeb4575857f3f2c4166e737c6b3863da62a2cfcf1cf",
					},
				},
			},
		},
	}
	_, err := chainDoc1.ToParsedDoc(make(map[string][]string))
	assert.ErrorContains(t, err, "failed to get content_group_label for content group: 0 in document with ID: 0, err: content group not found")

}

func TestToParsedDocShouldFailForInvalidInt(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						"d212",
					},
				},
			},
		},
	}
	_, err := chainDoc1.ToParsedDoc(make(map[string][]string))
	assert.ErrorContains(t, err, "failed to parse content value to int64")
}

func TestToParsedDocShouldFailForNoType(t *testing.T) {

	createdDate := "2020-11-12T18:27:47.000"
	updatedDate := "2020-11-12T20:37:47.000"
	chainDoc1 := &domain.ChainDocument{
		ID:          0,
		CreatedDate: createdDate,
		UpdatedDate: updatedDate,
		Creator:     "dao.hypha",
		Contract:    "contract1",
		ContentGroups: [][]*domain.ChainContent{
			{
				{
					Label: "content_group_label",
					Value: []interface{}{
						"name",
						"system",
					},
				},
				{
					Label: "votes",
					Value: []interface{}{
						"int64",
						"212",
					},
				},
			},
		},
	}
	_, err := chainDoc1.ToParsedDoc(make(map[string][]string))
	assert.ErrorContains(t, err, "document with ID: 0 does not have a type, and couldn't deduce from typeMappings")

	typeMappings := map[string][]string{
		"VoteTally": {
			"pass_votePower",
			"fail_votePower",
		},
	}
	_, err = chainDoc1.ToParsedDoc(typeMappings)
	assert.ErrorContains(t, err, "document with ID: 0 does not have a type, and couldn't deduce from typeMappings")
}

func TestChainDocUnmarshall(t *testing.T) {
	chainDocJSON := `{"content_groups":[[{"label":"content_group_label","value":["string","settings"]},{"label":"root_node","value":["string","52a7ff82bd6f53b31285e97d6806d886eefb650e79754784e9d923d3df347c91"]},{"label":"paused","value":["int64",0]},{"label":"updated_date","value":["time_point","2021-01-11T21:52:32"]},{"label":"seeds_token_contract","value":["name","token.seeds"]},{"label":"voting_duration_sec","value":["int64",3600]},{"label":"seeds_deferral_factor_x100","value":["int64",100]},{"label":"telos_decide_contract","value":["name","trailservice"]},{"label":"husd_token_contract","value":["name","husd.hypha"]},{"label":"hypha_token_contract","value":["name","token.hypha"]},{"label":"seeds_escrow_contract","value":["name","escrow.seeds"]},{"label":"publisher_contract","value":["name","publsh.hypha"]},{"label":"treasury_contract","value":["name","bank.hypha"]},{"label":"last_ballot_id","value":["name","hypha1....1cf"]},{"label":"hypha_deferral_factor_x100","value":["int64",25]},{"label":"client_version","value":["string","0.2.0 pre-release"]},{"label":"contract_version","value":["string","0.2.0 pre-release"]}],[{"label":"content_group_label","value":["string","system"]},{"label":"type","value":["name","settings"]},{"label":"node_label","value":["string","Settings"]}]],"contract":"dao.hypha","created_date":"2021-01-11T21:52:32","updated_date":"2021-01-11T22:52:32","creator":"dao.hypha","contract":"contract1","id":4957}`
	chainDoc := &domain.ChainDocument{}
	err := json.Unmarshal([]byte(chainDocJSON), chainDoc)
	if err != nil {
		t.Fatalf("Unmarshalling failed: %v", err)
	}
}

func assertParsedDoc(t *testing.T, actual, expected *domain.ParsedDoc) {
	util.AssertSimplifiedInstance(t, actual.Instance, expected.Instance)
	assert.DeepEqual(t, actual.ChecksumFields, expected.ChecksumFields)
}
