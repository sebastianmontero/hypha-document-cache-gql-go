package domain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
)

const CGL_ContentGroup = "content_group_label"

const CL_type = "system_type_n"

const (
	ContentType_Asset       = "asset"
	ContentType_Checksum256 = "checksum256"
	ContentType_Int64       = "int64"
	ContentType_Name        = "name"
	ContentType_Time        = "time_point"
	ContentType_String      = "string"
)

var ContentTypeGQLTypeMap = map[string]string{
	ContentType_Asset:       gql.GQLType_String,
	ContentType_Checksum256: gql.GQLType_String,
	ContentType_Int64:       gql.GQLType_Int64,
	ContentType_Name:        gql.GQLType_String,
	ContentType_Time:        gql.GQLType_Time,
	ContentType_String:      gql.GQLType_String,
}

var ContentTypeIndexMap = map[string]string{
	ContentType_Asset:       "term",
	ContentType_Checksum256: "exact",
	ContentType_Int64:       "int64",
	ContentType_Name:        "exact",
	ContentType_Time:        "hour",
	ContentType_String:      "regexp",
}

var ContentTypeSuffixMap = map[string]string{
	ContentType_Asset:       "a",
	ContentType_Checksum256: "c",
	ContentType_Int64:       "i",
	ContentType_Name:        "n",
	ContentType_Time:        "t",
	ContentType_String:      "s",
}

const CoreEdgeSuffix = "edge"

type ParsedDoc struct {
	Instance       *gql.SimplifiedInstance
	ChecksumFields []string
}

func (m *ParsedDoc) GetValue(name string) interface{} {
	return m.Instance.GetValue(name)
}

func (m *ParsedDoc) HasCoreEdges() bool {
	return m.NumCoreEdges() > 0
}

func (m *ParsedDoc) NumCoreEdges() int {
	return len(m.ChecksumFields)
}

//ChainDocs helper to enable chain docs decoding
type ChainDocs struct {
	Docs []*ChainDocument `json:"docs,omitempty"`
}

//ChainContent domain object
type ChainContent struct {
	Label string        `json:"label,omitempty"`
	Value []interface{} `json:"value,omitempty"`
}

func (m *ChainContent) GetType() string {
	return m.Value[0].(string)
}

func (m *ChainContent) IsChecksum() bool {
	return m.Value[0].(string) == ContentType_Checksum256
}

func (m *ChainContent) GetGQLType() string {
	return GetGQLType(m.GetType())
}

func (m *ChainContent) GetValue() string {
	return fmt.Sprintf("%v", m.Value[1])
}

func (m *ChainContent) GetGQLValue() (interface{}, error) {
	gqlType := m.GetGQLType()

	if gqlType == gql.GQLType_Time {
		return FormatDateTime(m.GetValue()), nil
	} else if gqlType == gql.GQLType_Int64 {
		intValue, err := strconv.ParseInt(m.GetValue(), 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content value to int64, value: %v for label: %v, error: %v", m.GetValue(), m.Label, err)
		}
		return intValue, nil
	} else {
		return m.GetValue(), nil
	}
}

func (m *ChainContent) String() string {
	return fmt.Sprintf("ChainContent{Label: %v, Value: %v}", m.Label, m.Value)
}

//ChainDocument domain object
type ChainDocument struct {
	ID            uint64            `json:"id"`
	CreatedDate   string            `json:"created_date,omitempty"`
	UpdatedDate   string            `json:"updated_date,omitempty"`
	Creator       string            `json:"creator,omitempty"`
	Contract      string            `json:"contract,omitempty"`
	ContentGroups [][]*ChainContent `json:"content_groups,omitempty"`
}

func (m *ChainDocument) ToParsedDoc(typeMappings map[string][]string) (*ParsedDoc, error) {

	fields := make(map[string]*gql.SimplifiedField)
	checksumFields := make([]string, 0)
	values := map[string]interface{}{
		"docId":       strconv.FormatUint(m.ID, 10),
		"docId_i":     m.ID,
		"creator":     m.Creator,
		"createdDate": FormatDateTime(m.CreatedDate),
		"updatedDate": FormatDateTime(m.UpdatedDate),
		"contract":    m.Contract,
	}

	for i, contentGroup := range m.ContentGroups {
		contentGroupLabel, err := GetContentGroupLabel(contentGroup)
		if err != nil {
			return nil, fmt.Errorf("failed to get content_group_label for content group: %v in document with ID: %v, err: %v", i, m.ID, err)
		}
		prefix := GetFieldPrefix(contentGroupLabel)
		for _, content := range contentGroup {
			if content.Label != CGL_ContentGroup {
				name := GetFieldName(prefix, content.Label, content.GetType())
				fields[name] = &gql.SimplifiedField{
					Name:  name,
					Type:  content.GetGQLType(),
					Index: GetIndex(content.GetType()),
				}
				if content.IsChecksum() {
					checksumFields = append(checksumFields, name)
				}
				value, err := content.GetGQLValue()
				if err != nil {
					return nil, fmt.Errorf("failed to get gql value content: %v name for doc with ID: %v, error: %v", name, m.ID, err)
				}
				values[name] = value
			}
		}
	}
	typeName, ok := values[CL_type].(string)
	if !ok {
		typeName = deduceDocType(toUntypedMap(fields), typeMappings)
		if typeName == "" {
			return nil, fmt.Errorf("document with ID: %v does not have a type, and couldn't deduce from typeMappings", m.ID)
		}
	}

	typeName = GetObjectTypeName(typeName)
	delete(values, CL_type)
	delete(fields, CL_type)
	values["type"] = typeName
	instance := gql.NewSimplifiedInstance(
		gql.NewSimplifiedType(typeName, fields, gql.DocumentSimplifiedInterface),
		values,
	)
	return &ParsedDoc{
		Instance:       instance,
		ChecksumFields: checksumFields,
	}, nil
}

func GetFieldPrefix(contentGroupLabel string) string {
	return fmt.Sprintf("%v", strcase.ToLowerCamel(contentGroupLabel))
}

func GetFieldName(cgPrefix, fieldLabel, fieldType string) string {
	return fmt.Sprintf("%v_%v_%v", cgPrefix, strcase.ToLowerCamel(fieldLabel), ContentTypeSuffixMap[fieldType])
}

func GetObjectTypeName(typeName string) string {
	return strcase.ToCamel(strings.ReplaceAll(typeName, ".", "_"))
}

func GetCoreEdgeName(checksumFieldName string) string {
	return fmt.Sprintf("%v_%v", checksumFieldName, CoreEdgeSuffix)
}

func GetGQLType(typeName string) string {
	return ContentTypeGQLTypeMap[typeName]
}

func GetIndex(typeName string) string {
	return ContentTypeIndexMap[typeName]
}

func IsBaseType(typeName string) bool {
	return typeName == ContentType_Asset || typeName == ContentType_Checksum256 ||
		typeName == ContentType_Int64 || typeName == ContentType_Name ||
		typeName == ContentType_Time || typeName == ContentType_String
}

func IsIDableType(typeName string) bool {
	return typeName == ContentType_Checksum256 ||
		typeName == ContentType_Name ||
		typeName == ContentType_String
}

func FindChainContent(contents []*ChainContent, label string) *ChainContent {
	for _, content := range contents {
		if content.Label == label {
			return content
		}
	}
	return nil
}

func GetContentGroupLabel(contents []*ChainContent) (string, error) {
	contentGroupLabel := FindChainContent(contents, CGL_ContentGroup)
	if contentGroupLabel == nil {
		return "", fmt.Errorf("content group not found")
	}
	return contentGroupLabel.GetValue(), nil
}

func deduceDocType(contentMap map[string]*gql.SimplifiedField, typeMappings map[string][]string) string {
	for typeName, labels := range typeMappings {
		if containsLabels(contentMap, labels) {
			return typeName
		}
	}
	return ""
}
func containsLabels(contentMap map[string]*gql.SimplifiedField, labels []string) bool {
	for _, label := range labels {
		if _, ok := contentMap[label]; !ok {
			return false
		}
	}
	return true
}

func toUntypedMap(typed map[string]*gql.SimplifiedField) map[string]*gql.SimplifiedField {
	untyped := make(map[string]*gql.SimplifiedField, len(typed))
	for label, value := range typed {
		pos := strings.LastIndex(label, "_")
		untypedLabel := label
		if pos > 0 {
			untypedLabel = string([]rune(label)[:pos])
		}
		untyped[untypedLabel] = value
	}
	return untyped
}

func (m *ChainDocument) String() string {
	return fmt.Sprintf("ChainDocument{ID: %v, CreatedDate: %v, UpdatedDate: %v, Creator: %v, Contents: %v}", m.ID, m.CreatedDate, m.UpdatedDate, m.Creator, m.ContentGroups)
}

func FormatDateTime(datetime string) string {
	return fmt.Sprintf("%vZ", datetime)
}
