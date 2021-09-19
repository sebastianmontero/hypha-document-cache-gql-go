package gql_test

import (
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/test/util"
	"gotest.tools/assert"
)

func TestApplyInterfacesNoApplicableInterface(t *testing.T) {
	interfaces := getMockInterfaces()
	var dhoType = gql.NewSimplifiedType(
		"Dho",
		map[string]*gql.SimplifiedField{
			"details_dho_n": {
				Name:  "details_dho_n",
				Type:  "String",
				Index: "exact",
			},
			"details_description_s": {
				Name:  "details_description_s",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(dhoType, nil)
	assert.NilError(t, err)

	expectedDhoType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Dho",
			Fields: map[string]*gql.SimplifiedField{
				"details_dho_n": {
					Name:  "details_dho_n",
					Type:  "String",
					Index: "exact",
				},
				"details_description_s": {
					Name:  "details_description_s",
					Type:  gql.GQLType_String,
					Index: "exact",
				},
			},
		},
	}
	util.AssertSimplifiedType(t, dhoType, expectedDhoType)
}

func TestApplyInterfacesSingleSignatureField(t *testing.T) {
	interfaces := getMockInterfaces()
	var memberType = gql.NewSimplifiedType(
		"Member",
		map[string]*gql.SimplifiedField{
			"details_account_n": {
				Name:  "details_account_n",
				Type:  "String",
				Index: "exact",
			},
			"details_profile_c": {
				Name:  "details_profile_c",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"memberName": {
				Name:  "memberName",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(memberType, nil)
	assert.NilError(t, err)

	expectedMemberType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "Member",
			Fields: map[string]*gql.SimplifiedField{
				"details_account_n": {
					Name:  "details_account_n",
					Type:  "String",
					Index: "exact",
				},
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
		},
		Interfaces: []string{"User"},
	}
	util.AssertSimplifiedType(t, memberType, expectedMemberType)
}

func TestApplyInterfacesMultipleSignatureFields(t *testing.T) {
	interfaces := getMockInterfaces()
	var badgeType = gql.NewSimplifiedType(
		"BadgeProposal",
		map[string]*gql.SimplifiedField{
			"details_badge_n": {
				Name:  "details_badge_n",
				Type:  "String",
				Index: "exact",
			},
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
			"ballot_expiration_t": {
				Name:  "ballot_expiration_t",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
			"votetally": {
				Name:    "votetally",
				Type:    "VoteTally",
				IsArray: true,
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(badgeType, nil)
	assert.NilError(t, err)

	expectedBadgeType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "BadgeProposal",
			Fields: map[string]*gql.SimplifiedField{
				"ballot_expiration_t": {
					Name:  "ballot_expiration_t",
					Type:  gql.GQLType_Time,
					Index: "hour",
				},
				"details_badge_n": {
					Name:  "details_badge_n",
					Type:  "String",
					Index: "exact",
				},
				"details_title_s": {
					Name:    "details_title_s",
					Type:    gql.GQLType_String,
					Index:   "regexp",
					IsID:    true,
					NonNull: true,
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
					Type:    "Document",
					IsArray: true,
				},
			},
		},
		Interfaces: []string{"Votable"},
	}
	util.AssertSimplifiedType(t, badgeType, expectedBadgeType)
}

func TestApplyInterfacesMultipleInterfaces(t *testing.T) {
	interfaces := getMockInterfaces()
	var memberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"details_account_n": {
				Name:  "details_badge_n",
				Type:  "String",
				Index: "exact",
			},
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
			"system_hash_c": {
				Name:  "system_hash_c",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"ballot_expiration_t": {
				Name:  "ballot_expiration_t",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
			"votetally": {
				Name:    "votetally",
				Type:    "VoteTally",
				IsArray: true,
			},
			"memberName": {
				Name:  "memberName",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"details_profile_c_edge": {
				Name: "details_profile_c_edge",
				Type: "ProfileData",
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(memberProposalType, nil)
	assert.NilError(t, err)

	expectedMemberProposalType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "MemberProposal",
			Fields: map[string]*gql.SimplifiedField{
				"details_account_n": {
					Name:  "details_badge_n",
					Type:  "String",
					Index: "exact",
				},
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
					Type:    "Document",
					IsArray: true,
				},
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
		},
		Interfaces: []string{"Votable", "User"},
	}
	util.AssertSimplifiedType(t, memberProposalType, expectedMemberProposalType)
}

func TestApplyInterfacesAddInterfaceToOldType(t *testing.T) {
	interfaces := getMockInterfaces()

	var oldMemberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
		},
		nil,
	)

	var memberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"memberName": {
				Name:  "memberName",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"details_profile_c_edge": {
				Name: "details_profile_c_edge",
				Type: "ProfileData",
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(memberProposalType, oldMemberProposalType)
	assert.NilError(t, err)

	expectedMemberProposalType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "MemberProposal",
			Fields: map[string]*gql.SimplifiedField{
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
		},
		Interfaces: []string{"User"},
	}
	util.AssertSimplifiedType(t, memberProposalType, expectedMemberProposalType)
}

func TestApplyInterfacesBasedOnOldTypeAndSignatureFields(t *testing.T) {
	interfaces := getMockInterfaces()

	var oldMemberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
		},
		interfaces["Votable"],
	)

	var memberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"details_account_n": {
				Name:  "details_badge_n",
				Type:  "String",
				Index: "exact",
			},
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
			"system_hash_c": {
				Name:  "system_hash_c",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"votetally": {
				Name:    "votetally",
				Type:    "VoteTally",
				IsArray: true,
			},
			"memberName": {
				Name:  "memberName",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"details_profile_c_edge": {
				Name: "details_profile_c_edge",
				Type: "ProfileData",
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(memberProposalType, oldMemberProposalType)
	assert.NilError(t, err)

	expectedMemberProposalType := &gql.SimplifiedType{
		SimplifiedBaseType: &gql.SimplifiedBaseType{
			Name: "MemberProposal",
			Fields: map[string]*gql.SimplifiedField{
				"details_account_n": {
					Name:  "details_badge_n",
					Type:  "String",
					Index: "exact",
				},
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
					Type:    "Document",
					IsArray: true,
				},
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
		},
		Interfaces: []string{"Votable", "User"},
	}
	util.AssertSimplifiedType(t, memberProposalType, expectedMemberProposalType)
}

func TestApplyInterfacesShouldFailForAddingIDFieldToOldType(t *testing.T) {
	interfaces := getMockInterfaces()

	var oldMemberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"details_account_n": {
				Name:  "details_badge_n",
				Type:  "String",
				Index: "exact",
			},
		},
		interfaces["Votable"],
	)

	var memberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"details_account_n": {
				Name:  "details_badge_n",
				Type:  "String",
				Index: "exact",
			},
			"system_hash_c": {
				Name:  "system_hash_c",
				Type:  gql.GQLType_String,
				Index: "exact",
			},
			"votetally": {
				Name:    "votetally",
				Type:    "VoteTally",
				IsArray: true,
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(memberProposalType, oldMemberProposalType)
	assert.ErrorContains(t, err, "can't add non null field")

}

func TestApplyInterfacesShouldFailForIncompatibleTypes(t *testing.T) {
	interfaces := getMockInterfaces()
	memberProposalType := gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"ballot_expiration_t": {
				Name:  "ballot_expiration_t",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
			"vote": {
				Name:    "vote",
				Type:    "VoteTally",
				IsArray: true,
			},
			"votetally": {
				Name:    "votetally",
				Type:    "Document",
				IsArray: true,
			},
		},
		nil,
	)

	err := interfaces.ApplyInterfaces(memberProposalType, nil)
	assert.ErrorContains(t, err, "can't make array field: vote of type: VoteTally, array of type: Vote")

	memberProposalType = gql.NewSimplifiedType(
		"MemberProposal",
		map[string]*gql.SimplifiedField{
			"ballot_expiration_t": {
				Name:  "ballot_expiration_t",
				Type:  gql.GQLType_Time,
				Index: "hour",
			},
			"details_title_s": {
				Name:  "details_title_s",
				Type:  gql.GQLType_String,
				Index: "regexp",
			},
			"vote": {
				Name: "vote",
				Type: "Vote",
			},
			"votetally": {
				Name:    "votetally",
				Type:    "Document",
				IsArray: true,
			},
		},
		nil,
	)

	err = interfaces.ApplyInterfaces(memberProposalType, nil)
	assert.ErrorContains(t, err, "can't make scalar field: vote an array")

}

func getMockInterfaces() gql.SimplifiedInterfaces {
	interfaces := gql.NewSimplifiedInterfaces()
	interfaces.Put(
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
					Type:    "Document",
					IsArray: true,
				},
			},
			[]string{
				"ballot_expiration_t",
				"votetally",
			},
		))
	interfaces.Put(
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
			},
		))
	return interfaces
}
