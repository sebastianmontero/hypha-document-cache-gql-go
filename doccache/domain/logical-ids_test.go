package domain_test

import (
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

func TestConfigureLogicalIdsNone(t *testing.T) {
	logicalIds := getMockLogicalIds()

	var assignmentType = gql.NewSimplifiedType(
		"Assignment",
		map[string]*gql.SimplifiedField{
			"system_hash_c": {
				Name:    "system_hash_c",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("exact"),
			},
			"details_name_n": {
				Name:    "details_name_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact"),
			},
		},
		nil,
	)

	err := logicalIds.ConfigureLogicalIds(assignmentType.SimplifiedBaseType)
	assert.NilError(t, err)

	expectedAssignmentType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Assignment",
			Fields: map[string]*gql.SimplifiedField{
				"details_name_n": {
					Name:    "details_name_n",
					Type:    "String",
					Indexes: gql.NewIndexes("exact"),
				},
				"system_hash_c": {
					Name:    "system_hash_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},
			},
			WithSubscription: true,
		},
	}
	util.AssertSimplifiedType(t, assignmentType, expectedAssignmentType)

}

func TestConfigureLogicalIdsSingleId(t *testing.T) {
	logicalIds := getMockLogicalIds()

	var dhoType = gql.NewSimplifiedType(
		"Dho",
		map[string]*gql.SimplifiedField{
			"system_hash_c": {
				Name:    "system_hash_c",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("exact"),
			},
			"details_name_n": {
				Name:    "details_name_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact"),
			},
		},
		nil,
	)

	err := logicalIds.ConfigureLogicalIds(dhoType.SimplifiedBaseType)
	assert.NilError(t, err)

	expectedDhoType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Dho",
			Fields: map[string]*gql.SimplifiedField{
				"details_name_n": {
					Name:    "details_name_n",
					Type:    "String",
					Indexes: gql.NewIndexes("exact"),
					IsID:    true,
					NonNull: true,
				},
				"system_hash_c": {
					Name:    "system_hash_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},
			},
			WithSubscription: true,
		},
	}
	util.AssertSimplifiedType(t, dhoType, expectedDhoType)

}
func TestConfigureLogicalIdsMultipleIds(t *testing.T) {
	logicalIds := getMockLogicalIds()

	var memberType = gql.NewSimplifiedType(
		"Member",
		map[string]*gql.SimplifiedField{
			"details_member_n": {
				Name:    "details_member_n",
				Type:    "String",
				Indexes: gql.NewIndexes("exact"),
			},
			"system_hash_c": {
				Name:    "system_hash_c",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("exact"),
			},
			"memberName": {
				Name:    "memberName",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("exact"),
			},
		},
		nil,
	)

	err := logicalIds.ConfigureLogicalIds(memberType.SimplifiedBaseType)
	assert.NilError(t, err)

	expectedMemberType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Member",
			Fields: map[string]*gql.SimplifiedField{
				"details_member_n": {
					Name:    "details_member_n",
					Type:    "String",
					Indexes: gql.NewIndexes("exact"),
					IsID:    true,
					NonNull: true,
				},
				"system_hash_c": {
					Name:    "system_hash_c",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
					IsID:    true,
					NonNull: true,
				},
				"memberName": {
					Name:    "memberName",
					Type:    gql.GQLType_String,
					Indexes: gql.NewIndexes("exact"),
				},
			},
			WithSubscription: true,
		},
	}
	util.AssertSimplifiedType(t, memberType, expectedMemberType)

}

func TestConfigureLogicalShouldFailForMissingIdField(t *testing.T) {
	logicalIds := getMockLogicalIds()

	var dhoType = gql.NewSimplifiedType(
		"Dho",
		map[string]*gql.SimplifiedField{
			"system_hash_c": {
				Name:    "system_hash_c",
				Type:    gql.GQLType_String,
				Indexes: gql.NewIndexes("exact"),
			},
		},
		nil,
	)

	err := logicalIds.ConfigureLogicalIds(dhoType.SimplifiedBaseType)
	assert.ErrorContains(t, err, "failed configuring logical ids")

}

func getMockLogicalIds() domain.LogicalIds {
	logicalIds := domain.NewLogicalIds()
	logicalIds.Set(
		"Member",
		[]string{
			"details_member_n",
			"system_hash_c",
		},
	)
	logicalIds.Set(
		"Dho",
		[]string{
			"details_name_n",
		},
	)
	return logicalIds
}
