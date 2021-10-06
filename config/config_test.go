package config_test

import (
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/config"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

func TestOptionalConfigsNotSet(t *testing.T) {
	config, err := config.LoadConfig("./config-optionals-nil.yml")
	assert.NilError(t, err)

	assert.Assert(t, config.TypeMappings == nil)
	assert.Assert(t, config.Interfaces == nil)
	assert.Assert(t, config.LogicalIds == nil)
	assert.Equal(t, len(config.TypeMappings), 0)
	assert.Equal(t, len(config.Interfaces), 0)
	assert.Equal(t, len(config.LogicalIds), 0)
}

func TestLoadTypeMappings(t *testing.T) {
	config, err := config.LoadConfig("./config-type-mappings.yml")
	assert.NilError(t, err)
	expected := map[string][]string{
		"VoteTally": {
			"pass_votePower",
			"fail_votePower",
		},
		"Assignment": {
			"details_assignment",
		},
	}
	AssertTypeMappings(t, config.TypeMappings, expected)

}

func TestLoadInterfaces(t *testing.T) {
	config, err := config.LoadConfig("./config-interfaces.yml")
	assert.NilError(t, err)
	expected := gql.NewSimplifiedInterfaces()
	expected.Put(
		gql.NewSimplifiedInterface(
			"Votable",
			map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:  "ballot_expiration_t",
					Type:  gql.GQLType_Time,
					Index: "hour",
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Index:   "regexp",
					IsID:    true,
					NonNull: true,
				},
				"details_description_s": {
					Name:  "details_description_s",
					Type:  gql.GQLType_String,
					Index: "regexp",
				},
				"system_hash_c": {
					Name:  "system_hash_c",
					Type:  gql.GQLType_String,
					Index: "exact",
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
			nil,
		))
	expected.Put(
		gql.NewSimplifiedInterface(
			"User",
			map[string]*gql.SimplifiedField{
				"details_profile_c": {
					Name:  "details_profile_c",
					Type:  gql.GQLType_String,
					Index: "exact",
				},
				"details_profile_c_edge": {
					Name: "details_profile_c_edge",
					Type: "ProfileData",
				},
				"memberName": {
					Name:  "memberName",
					Type:  gql.GQLType_String,
					Index: "exact",
				},
			},
			[]string{
				"memberName",
				"details_profile_c",
			},
			[]string{
				"Owner",
				"Admin",
			},
		))
	expected.Put(
		gql.NewSimplifiedInterface(
			"Editable",
			map[string]*gql.SimplifiedField{
				"details_version_s": {
					Name:  "details_version_s",
					Type:  gql.GQLType_String,
					Index: "regexp",
				},
			},
			nil,
			[]string{
				"ProPaper",
			},
		))

	util.AssertSimplifiedInterfaces(t, config.Interfaces, expected)

}

func TestLoadInterfacesShouldFailForNoSignatureFields(t *testing.T) {
	_, err := config.LoadConfig("./config-interfaces-no-signature-types.yml")
	assert.ErrorContains(t, err, "it must have at least one signature field or type specified")

}

func TestLoadInterfacesShouldFailForIdOnInvalidType(t *testing.T) {
	_, err := config.LoadConfig("./config-interfaces-invalid-id.yml")
	assert.ErrorContains(t, err, "id fields can only be of IDable types")

}

func TestLoadLogicalIds(t *testing.T) {
	config, err := config.LoadConfig("./config-logical-ids.yml")
	assert.NilError(t, err)
	expected := domain.NewLogicalIds()
	expected.Set(
		"Member",
		[]string{
			"details_member_n",
			"system_hash_c",
			"system_memberId_s",
		},
	)
	expected.Set(
		"Dho",
		[]string{
			"details_name_n",
		},
	)
	AssertLogicalIds(t, config.LogicalIds, expected)
}

func TestLoadLogicalIdsShouldFailForInvalidType(t *testing.T) {
	_, err := config.LoadConfig("./config-logical-ids-invalid-type.yml")
	assert.ErrorContains(t, err, "id fields can only be of IDable types")
}

func AssertTypeMappings(t *testing.T, actual, expected map[string][]string) {
	assert.Equal(t, len(actual), len(expected), "Different number of types actual: %v, expected: %v", actual, expected)
	for eName, eFields := range expected {
		aFields := actual[eName]
		util.AssertUnorderedStrArray(t, aFields, eFields)
	}
}

func AssertLogicalIds(t *testing.T, actual, expected map[string][]string) {
	assert.Equal(t, len(actual), len(expected), "Different number of types actual: %v, expected: %v", actual, expected)
	for eName, eFields := range expected {
		aFields := actual[eName]
		util.AssertUnorderedStrArray(t, aFields, eFields)
	}
}
