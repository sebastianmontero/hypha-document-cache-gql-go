package doccache

import (
	"fmt"
	"strconv"
	"time"
)

//Docs helper to enable docs decoding
type Docs struct {
	Docs []*Document `json:"docs,omitempty"`
}

//Content domain object
type Content struct {
	UID             string      `json:"uid,omitempty"`
	Label           string      `json:"label,omitempty"`
	Value           string      `json:"value,omitempty"`
	TimeValue       *time.Time  `json:"time_value,omitempty"`
	IntValue        *int64      `json:"int_value,omitempty"`
	Type            string      `json:"type,omitempty"`
	ContentSequence int         `json:"content_sequence"`
	Document        []*Document `json:"document,omitempty"`
	DType           []string    `json:"dgraph.type,omitempty"`
}

//NewContent Creates a Content object based on a ChainContent
func NewContent(chainContent *ChainContent, sequence int) (*Content, error) {
	content := &Content{
		Label:           chainContent.Label,
		Type:            fmt.Sprintf("%v", chainContent.Value[0]),
		Value:           fmt.Sprintf("%v", chainContent.Value[1]),
		ContentSequence: sequence,
		DType:           []string{"Content"},
	}
	if content.IsInt64() {
		intValue, err := strconv.ParseInt(content.Value, 0, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse content value to int64, value: %v, error: %v", content.Value, err)
		}
		content.IntValue = &intValue
	}
	if content.IsTime() {
		content.TimeValue = ToTime(content.Value)
	}
	return content, nil
}

//IsChecksum indicates if Content is of type Checksum
func (m *Content) IsChecksum() bool {
	return m.Type == "checksum256"
}

//IsInt64 indicates if Content is of type int64
func (m *Content) IsInt64() bool {
	return m.Type == "int64"
}

//IsTime indicates if Content is of type time
func (m *Content) IsTime() bool {
	return m.Type == "time_point"
}

func (m *Content) String() string {
	return fmt.Sprintf("Content{UID: %v, Label: %v, Value: %v, TimeValue: %v, IntValue: %v, ContentSequence: %v, Document: %v, DType: %v}", m.UID, m.Label, m.Value, m.TimeValue, m.IntValue, m.ContentSequence, m.Document, m.DType)
}

//ContentGroup domain object
type ContentGroup struct {
	UID                  string     `json:"uid,omitempty"`
	ContentGroupSequence int        `json:"content_group_sequence"`
	Contents             []*Content `json:"contents,omitempty"`
	DType                []string   `json:"dgraph.type,omitempty"`
}

//NewContentGroup Creates a ContentGroup based on a ChainContentGroup
func NewContentGroup(chainContentGroup []*ChainContent, sequence int) (*ContentGroup, error) {
	contents := make([]*Content, 0, len(chainContentGroup))
	for i, chainContent := range chainContentGroup {
		content, err := NewContent(chainContent, i+1)
		if err != nil {
			return nil, err
		}
		contents = append(contents, content)
	}
	return &ContentGroup{
		ContentGroupSequence: sequence,
		Contents:             contents,
		DType:                []string{"ContentGroup"},
	}, nil
}

//GetChecksumContents returns Contents with checksum type
func (m *ContentGroup) GetChecksumContents() []*Content {
	found := make([]*Content, 0)
	for _, content := range m.Contents {
		if content.IsChecksum() {
			found = append(found, content)
		}
	}
	return found
}

func (m *ContentGroup) String() string {
	return fmt.Sprintf("ContentGroup{UID: %v, ContentGroupSequence: %v, Contents: %v, DType: %v}", m.UID, m.ContentGroupSequence, m.Contents, m.DType)
}

//Certificate domain object
type Certificate struct {
	UID                   string     `json:"uid,omitempty"`
	Certifier             string     `json:"certifier,omitempty"`
	Notes                 string     `json:"notes,omitempty"`
	CertificationDate     *time.Time `json:"certification_date,omitempty"`
	CertificationSequence int        `json:"certification_sequence"`
	DType                 []string   `json:"dgraph.type,omitempty"`
}

//NewCertificate Creates a Certificate based on a ChainCertificate
func NewCertificate(chainCertificate *ChainCertificate, sequence int) *Certificate {
	return &Certificate{
		Certifier:             chainCertificate.Certifier,
		Notes:                 chainCertificate.Notes,
		CertificationDate:     ToTime(chainCertificate.CertificationDate),
		CertificationSequence: sequence,
		DType:                 []string{"Certificate"},
	}
}

func (m *Certificate) String() string {
	return fmt.Sprintf("Certificate{UID: %v, Certifier: %v, Notes: %v, CertificationDate: %v, CertificationSequence: %v, DType: %v}", m.UID, m.Certifier, m.Notes, m.CertificationDate, m.CertificationSequence, m.DType)
}

//Document domain object
type Document struct {
	UID           string          `json:"uid,omitempty"`
	Hash          string          `json:"hash,omitempty"`
	CreatedDate   *time.Time      `json:"created_date,omitempty"`
	Creator       string          `json:"creator,omitempty"`
	ContentGroups []*ContentGroup `json:"content_groups,omitempty"`
	Certificates  []*Certificate  `json:"certificates,omitempty"`
	DType         []string        `json:"dgraph.type,omitempty"`
}

//NewDocument creates a new document from a ChainDocument
func NewDocument(chainDoc *ChainDocument) (*Document, error) {
	contentGroups := make([]*ContentGroup, 0, len(chainDoc.ContentGroups))
	certificates := make([]*Certificate, 0, len(chainDoc.Certificates))

	for i, chainContentGroup := range chainDoc.ContentGroups {
		contentGroup, err := NewContentGroup(chainContentGroup, i+1)
		if err != nil {
			return nil, err
		}
		contentGroups = append(contentGroups, contentGroup)
	}

	for i, chainCertificate := range chainDoc.Certificates {
		certificates = append(certificates, NewCertificate(chainCertificate, i+1))
	}

	return &Document{
		Hash:          chainDoc.Hash,
		CreatedDate:   ToTime(chainDoc.CreatedDate),
		Creator:       chainDoc.Creator,
		ContentGroups: contentGroups,
		Certificates:  certificates,
		DType:         []string{"Document"},
	}, nil
}

//GetChecksumContents returns Contents with checksum type
func (m *Document) GetChecksumContents() []*Content {
	found := make([]*Content, 0)
	for _, contentGroup := range m.ContentGroups {
		found = append(found, contentGroup.GetChecksumContents()...)
	}
	return found
}

//UpdateCertificates updates doc certificates
func (m *Document) UpdateCertificates(chainCertificates []*ChainCertificate) {
	for i := len(m.Certificates); i < len(chainCertificates); i++ {
		m.Certificates = append(m.Certificates, NewCertificate(chainCertificates[i], i+1))
	}
}

func (m *Document) String() string {
	return fmt.Sprintf("Document{UID: %v, Hash: %v, CreatedDate: %v, Creator: %v, ContentGroups: %v, Certificates: %v, DType: %v}", m.UID, m.Hash, m.CreatedDate, m.Creator, m.ContentGroups, m.Certificates, m.DType)
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

// func (m *ChainContent) UnmarshalJSON(b []byte) error {
// 	if err := json.Unmarshal(b, m); err != nil {
// 		return err
// 	}
// 	if fmt.Sprintf("%v", m.Value[0]) == "checksum256" {
// 		m.Value[1] = strings.ToUpper(fmt.Sprintf("%v", m.Value[1]))
// 	}
// 	return nil
// }

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

// func (m *ChainDocument) UnmarshalJSON(b []byte) error {
// 	if err := json.Unmarshal(b, m); err != nil {
// 		return err
// 	}
// 	m.Hash = strings.ToUpper(m.Hash)
// 	return nil
// }

func (m *ChainDocument) String() string {
	return fmt.Sprintf("ChainDocument{ID: %v, Hash: %v, CreatedDate: %v, Creator: %v, Contents: %v, Certificates: %v}", m.ID, m.Hash, m.CreatedDate, m.Creator, m.ContentGroups, m.Certificates)
}

//ChainEdge domain object
type ChainEdge struct {
	Name string `json:"edge_name,omitempty"`
	From string `json:"from_node,omitempty"`
	To   string `json:"to_node,omitempty"`
}

// func (m *ChainEdge) UnmarshalJSON(b []byte) error {
// 	if err := json.Unmarshal(b, m); err != nil {
// 		return err
// 	}
// 	m.From = strings.ToUpper(m.From)
// 	m.To = strings.ToUpper(m.To)
// 	return nil
// }

func (m *ChainEdge) String() string {
	return fmt.Sprintf("ChainEdge{Name: %v, From: %v, To: %v}", m.Name, m.From, m.To)
}

//Cursors helper to enable cursor decoding
type Cursors struct {
	Cursors []*Cursor `json:"cursors,omitempty"`
}

//Cursor domain object
type Cursor struct {
	UID    string   `json:"uid,omitempty"`
	Cursor string   `json:"cursor,omitempty"`
	DType  []string `json:"dgraph.type,omitempty"`
}

func (m *Cursor) String() string {
	return fmt.Sprintf("Cursor{UID: %v, Cursor: %v, DType: %v}", m.UID, m.Cursor, m.DType)
}

//ToTime Converts string time to time.Time
func ToTime(strTime string) *time.Time {
	t, err := time.Parse("2006-01-02T15:04:05", strTime)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05.000", strTime)
		if err != nil {
			log.Errorf(err, "Failed to parse datetime: %v", strTime)
		}
	}
	return &t
}
