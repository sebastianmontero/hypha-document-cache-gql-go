package doccache

import (
	"fmt"
	"strings"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/slog-go/slog"
)

const schema = `
      type Document {
          hash
          created_date
          creator
          content_groups
          certificates
      }
      
      type ContentGroup {
        content_group_sequence
        contents
      }
      
      type Content {
        label
        value
        type
        content_sequence
        document
				time_value
				int_value
      }
      
      type Certificate {
        certifier
        notes
        certification_date
        certification_sequence
      }

			type Cursor {
				cursor
			}
      
      hash: string @index(exact) .
      created_date: datetime @index(hour) .
      creator: string @index(term) .
      content_groups: [uid] .
      certificates: [uid] .
      
      content_group_sequence: int .
      contents: [uid] .
      
      label: string @index(term) .
      value: string @index(term) .
      type: string @index(term) .
      content_sequence: int .
      document: [uid] .
			time_value: datetime @index(hour) .
			int_value: int @index(int) .
      
      certifier: string @index(term) .
      notes: string .
      certification_date: datetime @index(hour) .
      certification_sequence: int .

			cursor: string @index(term) .
    `
const contentGroupsRequest = `
      content_groups (orderasc:content_group_sequence){
				content_group_sequence
				dgraph.type
        contents (orderasc: content_sequence){
          content_sequence
          label
          value
					type
					time_value
					int_value
					dgraph.type
          document{
            expand(_all_)
          }
        }
      },
    `

const certificatesRequest = `
      certificates (orderasc: certification_sequence){
				uid
				dgraph.type
        expand(_all_)
      },
		`

var log *slog.Log

//RequestConfig enables query configuration
type RequestConfig struct {
	ContentGroups bool
	Certificates  bool
	Edges         []string
}

//Doccache Service class to store and retrieve docs
type Doccache struct {
	dgraph           *dgraph.Dgraph
	documentFieldMap map[string]*dgraph.SchemaField
	Cursor           *Cursor
}

//New creates a new doccache
func New(dg *dgraph.Dgraph, logConfig *slog.Config) (*Doccache, error) {
	log = slog.New(logConfig, "doccache")

	m := &Doccache{
		dgraph:           dg,
		documentFieldMap: make(map[string]*dgraph.SchemaField),
	}

	err := m.PrepareSchema()
	if err != nil {
		return nil, err
	}
	cursor, err := m.getCursor()
	if err != nil {
		return nil, err
	}
	m.Cursor = cursor
	return m, nil
}

//SchemaExists set the base document schema in dgraph
func (m *Doccache) SchemaExists() (bool, error) {
	missing, err := m.dgraph.MissingTypes([]string{"Document", "ContentGroup", "Content", "Certificate", "Cursor"})
	if err != nil {
		return false, err
	}
	return len(missing) == 0, nil
}

//PrepareSchema prepares the base schema
func (m *Doccache) PrepareSchema() error {
	exists, err := m.SchemaExists()
	if err != nil {
		return err
	}
	if !exists {
		err = m.dgraph.UpdateSchema(schema)
		if err != nil {
			return err
		}
	}
	m.documentFieldMap, err = m.dgraph.GetTypeFieldMap("Document")
	return err
}

//GetCursor Finds the current cursor
func (m *Doccache) getCursor() (*Cursor, error) {
	query := `
		{
			cursors(func: type(Cursor)){
				uid
				cursor
				dgraph.type
			}
		}
	`
	cursors := &Cursors{}
	err := m.dgraph.Query(query, nil, cursors)
	if err != nil {
		return nil, err
	}
	if len(cursors.Cursors) > 0 {
		return cursors.Cursors[0], nil
	}

	return m.createCursor()
}

func (m *Doccache) createCursor() (*Cursor, error) {
	cursor := &Cursor{
		DType: []string{"Cursor"},
	}
	mutation, err := m.cursorMutation(cursor)
	if err != nil {
		return nil, err
	}
	resp, err := m.dgraph.MutateOne(mutation)
	if err != nil {
		return nil, err
	}
	for _, uid := range resp.GetUids() {
		cursor.UID = uid
	}
	return cursor, nil
}

func (m *Doccache) cursorMutation(cursor *Cursor) (*api.Mutation, error) {
	return m.dgraph.JSONMutation(cursor, false)
}

//GetByHash Finds document by hash
func (m *Doccache) GetByHash(hash string, rc *RequestConfig) (*Document, error) {
	query := fmt.Sprintf(`
		query docs($hash: string){
			docs(func: eq(hash, $hash))
				%v
		}
	`, configureRequest(rc))

	docs := &Docs{}
	err := m.dgraph.Query(query, map[string]string{"$hash": hash}, docs)
	if err != nil {
		return nil, err
	}
	if len(docs.Docs) > 0 {
		return docs.Docs[0], nil
	}
	return nil, nil
}

//GetByHashAsMap Finds document by hash returns a map
func (m *Doccache) GetByHashAsMap(hash string, rc *RequestConfig) (map[string]interface{}, error) {
	query := fmt.Sprintf(`
		query docs($hash: string){
			docs(func: eq(hash, $hash))
				%v
		}
	`, configureRequest(rc))

	documents := make(map[string]interface{})
	err := m.dgraph.Query(query, map[string]string{"$hash": hash}, &documents)
	if err != nil {
		return nil, err
	}
	if docsi, ok := documents["docs"]; ok {
		docs := docsi.([]interface{})
		if len(docs) > 0 {
			return docs[0].(map[string]interface{}), nil
		}
		return nil, nil
	}
	return nil, nil
}

//GetHashUIDMap finds docs by hashes and returns a map hash->uid
func (m *Doccache) GetHashUIDMap(hashes []string) (map[string]string, error) {
	if len(hashes) == 0 {
		return make(map[string]string), nil
	}
	query := fmt.Sprintf(`
		{
			docs(func: eq(hash, [%v])){
				uid
				hash
			}
		}
	`, strings.Join(hashes, ","))

	docs := &Docs{}
	err := m.dgraph.Query(query, nil, docs)
	if err != nil {
		return nil, err
	}
	var hashUIDMap = make(map[string]string, len(hashes))

	for _, doc := range docs.Docs {
		hashUIDMap[doc.Hash] = doc.UID
	}
	return hashUIDMap, nil
}

//GetUID finds UID from hash
func (m *Doccache) GetUID(hash string) (string, error) {
	hashUIDMap, err := m.GetHashUIDMap([]string{hash})
	if err != nil {
		return "", err
	}
	if uid, ok := hashUIDMap[hash]; ok {
		return uid, nil
	}
	return "", nil
}

func (m *Doccache) mutate(mutation *api.Mutation, cursor string) error {
	m.Cursor.Cursor = cursor
	cursorMutation, err := m.cursorMutation(m.Cursor)
	if err != nil {
		return err
	}
	_, err = m.dgraph.Mutate(mutation, cursorMutation)
	return err
}

func (m *Doccache) UpdateCursor(cursor string) error {
	m.Cursor.Cursor = cursor
	cursorMutation, err := m.cursorMutation(m.Cursor)
	if err != nil {
		return err
	}
	_, err = m.dgraph.Mutate(cursorMutation)
	return err
}

//StoreDocument Creates a new document or updates its certificates
func (m *Doccache) StoreDocument(chainDoc *ChainDocument, cursor string) error {
	doc, err := m.GetByHash(chainDoc.Hash, &RequestConfig{Certificates: true})
	if err != nil {
		return err
	}
	if doc == nil {
		log.Infof("Creating document: %v", chainDoc.Hash)
		doc, err = m.transformNew(chainDoc)
		if err != nil {
			return err
		}
	} else {
		log.Infof("Updating certificates for document: <%v>%v", doc.UID, doc.Hash)
		doc.UpdateCertificates(chainDoc.Certificates)
	}

	mutation, err := m.dgraph.JSONMutation(doc, false)
	if err != nil {
		return err
	}
	return m.mutate(mutation, cursor)
}

//DeleteDocument Deletes a document
func (m *Doccache) DeleteDocument(chainDoc *ChainDocument, cursor string) error {
	uid, err := m.GetUID(chainDoc.Hash)
	if err != nil {
		return err
	}
	if uid != "" {
		log.Infof("Deleting Node: <%v>%v", uid, chainDoc.Hash)
		mutation := m.dgraph.DeleteNodeMutation(uid)
		return m.mutate(mutation, cursor)
	}
	log.Infof("Document: %v not found, couldn't delete", chainDoc.Hash)
	return nil
}

//MutateEdge Creates/Deletes an edge
func (m *Doccache) MutateEdge(chainEdge *ChainEdge, deleteOp bool, cursor string) error {
	err := m.updateDocumentTypeSchema(chainEdge.Name)
	if err != nil {
		return err
	}
	hashUIDMap, err := m.GetHashUIDMap([]string{chainEdge.From, chainEdge.To})
	if err != nil {
		return err
	}
	fromUID, ok := hashUIDMap[chainEdge.From]
	if !ok {
		return fmt.Errorf("from node of the relationship: [Edge: %v, From: %v, To: %v] does not exist, Delete Op: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp)
	}
	toUID, ok := hashUIDMap[chainEdge.To]
	if !ok {
		return fmt.Errorf("to node of the relationship: [Edge: %v, From: %v, To: %v] does not exist, Delete Op: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp)
	}
	log.Infof("Mutating [Edge: %v, From: <%v>%v, To: <%v>%v] Delete Op: %v", chainEdge.Name, fromUID, chainEdge.From, toUID, chainEdge.To, deleteOp)
	mutation := m.dgraph.EdgeMutation(fromUID, toUID, chainEdge.Name, deleteOp)
	return m.mutate(mutation, cursor)

}

func (m *Doccache) updateDocumentTypeSchema(newField string) error {
	if _, ok := m.documentFieldMap[newField]; !ok {
		fields := ""
		for key := range m.documentFieldMap {
			fields += "\n" + key
		}
		err := m.dgraph.UpdateSchema(fmt.Sprintf(
			`
				%v: [uid] .
				type Document{
					%v
					%v
				}
		 `, newField, fields, newField))
		if err != nil {
			return err
		}
		m.documentFieldMap[newField] = &dgraph.SchemaField{Name: newField}
	}
	return nil
}

func (m *Doccache) transformNew(chainDoc *ChainDocument) (*Document, error) {
	doc, err := NewDocument(chainDoc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain doc: %v \n error: %v", chainDoc, err)
	}
	checksumContents := doc.GetChecksumContents()
	hashes := make([]string, 0, len(checksumContents))
	for _, checksumContent := range checksumContents {
		hashes = append(hashes, checksumContent.Value)
	}
	hashUIDMap, err := m.GetHashUIDMap(hashes)

	if err != nil {
		return nil, err
	}

	for _, checksumContent := range checksumContents {
		if uid, ok := hashUIDMap[checksumContent.Value]; ok {
			checksumContent.Document = []*Document{
				{
					UID: uid,
				},
			}
		} else {
			log.Infof("Document with hash: %v not found, referenced from document: %v", checksumContent.Value, chainDoc.Hash)
		}
	}
	return doc, nil
}

func configureRequest(rc *RequestConfig) string {
	contentGroups, certificates := "", ""
	if rc.ContentGroups {
		contentGroups = contentGroupsRequest
	}
	if rc.Certificates {
		certificates = certificatesRequest
	}
	predicates := fmt.Sprintf(`
		uid
		hash
		creator
		created_date
		dgraph.type
		%v
		%v
	`, contentGroups, certificates)

	edgeRequest := ""

	for _, edge := range rc.Edges {
		edgeRequest += fmt.Sprintf(`
			%v {
				%v
			}
		`, edge, predicates)
	}
	return fmt.Sprintf(`
		{
			%v,
			%v
		}
	`, predicates, edgeRequest)
}
