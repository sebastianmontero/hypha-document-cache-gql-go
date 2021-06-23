package domain

import (
	"fmt"
	"strconv"

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
	return ContentTypeGQLTypeMap[m.GetType()]
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

//ChainCertificate domain object
type ChainCertificate struct {
	Certifier         string `json:"certifier,omitempty"`
	Notes             string `json:"notes,omitempty"`
	CertificationDate string `json:"certification_date,omitempty"`
}

func (m *ChainCertificate) String() string {
	return fmt.Sprintf("ChainCertificate{Certifier: %v, Notes: %v, CertificationDate: %v}", m.Certifier, m.Notes, m.CertificationDate)
}

//ChainDocument domain object
type ChainDocument struct {
	ID            int                 `json:"id"`
	Hash          string              `json:"hash,omitempty"`
	CreatedDate   string              `json:"created_date,omitempty"`
	Creator       string              `json:"creator,omitempty"`
	ContentGroups [][]*ChainContent   `json:"content_groups,omitempty"`
	Certificates  []*ChainCertificate `json:"certificates,omitempty"`
}

func (m *ChainDocument) ToParsedDoc() (*ParsedDoc, error) {

	fields := make(map[string]*gql.SimplifiedField)
	checksumFields := make([]string, 0)
	values := map[string]interface{}{
		"hash":        m.Hash,
		"creator":     m.Creator,
		"createdDate": FormatDateTime(m.CreatedDate),
	}

	for i, contentGroup := range m.ContentGroups {
		contentGroupLabel := FindChainContent(contentGroup, CGL_ContentGroup)
		if contentGroupLabel == nil {
			return nil, fmt.Errorf("content group: %v for document with hash: %v does not have a content_group_label", i, m.Hash)
		}
		prefix := fmt.Sprintf("%v", strcase.ToLowerCamel(contentGroupLabel.GetValue()))
		for _, content := range contentGroup {
			if content.Label != CGL_ContentGroup {
				name := fmt.Sprintf("%v_%v_%v", prefix, strcase.ToLowerCamel(content.Label), ContentTypeSuffixMap[content.GetType()])
				fields[name] = &gql.SimplifiedField{
					Name:  name,
					Type:  content.GetGQLType(),
					Index: ContentTypeIndexMap[content.GetType()],
				}
				if content.IsChecksum() {
					checksumFields = append(checksumFields, name)
				}
				value, err := content.GetGQLValue()
				if err != nil {
					return nil, fmt.Errorf("failed to get gql value content: %v name for doc with hash: %v, error: %v", name, m.Hash, err)
				}
				values[name] = value
			}
		}
	}
	typeName, ok := values[CL_type].(string)
	if !ok {
		return nil, fmt.Errorf("document with hash: %v does not have a type", m.Hash)
	}
	typeName = strcase.ToCamel(typeName)
	delete(values, CL_type)
	delete(fields, CL_type)
	values["type"] = typeName
	instance := &gql.SimplifiedInstance{
		SimplifiedType: &gql.SimplifiedType{
			Name:            typeName,
			Fields:          fields,
			ExtendsDocument: true,
		},
		Values: values,
	}
	return &ParsedDoc{
		Instance:       instance,
		ChecksumFields: checksumFields,
	}, nil
}

func FindChainContent(contents []*ChainContent, label string) *ChainContent {
	for _, content := range contents {
		if content.Label == label {
			return content
		}
	}
	return nil
}

func (m *ChainDocument) String() string {
	return fmt.Sprintf("ChainDocument{ID: %v, Hash: %v, CreatedDate: %v, Creator: %v, Contents: %v, Certificates: %v}", m.ID, m.Hash, m.CreatedDate, m.Creator, m.ContentGroups, m.Certificates)
}

func FormatDateTime(datetime string) string {
	return fmt.Sprintf("%vZ", datetime)
}
