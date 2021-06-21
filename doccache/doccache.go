package doccache

import (
	"fmt"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/slog-go/slog"
)

const CoreEdgeSuffix = "edge"
const CursorId string = "c1"

var log *slog.Log

//Doccache Service class to store and retrieve docs
type Doccache struct {
	dgraph *dgraph.Dgraph
	admin  *gql.Admin
	client *gql.Client
	Cursor *gql.SimplifiedInstance
	Schema *gql.Schema
}

//New creates a new doccache
func New(dg *dgraph.Dgraph, admin *gql.Admin, client *gql.Client, logConfig *slog.Config) (*Doccache, error) {
	log = slog.New(logConfig, "doccache")

	m := &Doccache{
		dgraph: dg,
		admin:  admin,
		client: client,
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
	log.Infof("Getting current schema...")
	schema, err := m.admin.GetCurrentSchema()
	fmt.Println("Current schema: ", schema)
	if err != nil {
		return fmt.Errorf("failed getting current schema, error: %v", err)
	}
	if schema == nil {
		log.Infof("No current schema, initializing schema...")
		schema, err = gql.InitialSchema()
		if err != nil {
			return fmt.Errorf("failed getting initial schema, error: %v", err)
		}
		err = m.admin.UpdateSchema(schema)
		if err != nil {
			return fmt.Errorf("failed setting initial schema, error: %v", err)
		}
		log.Infof("Initialized schema.")
	}
	m.Schema = schema
	return nil
}

//GetCursor Finds the current cursor
func (m *Doccache) getCursor() (*gql.SimplifiedInstance, error) {

	cursor, err := m.client.GetOne(CursorId, gql.CursorSimplifiedType, nil)
	if err != nil {
		return nil, fmt.Errorf("failed getting cursor with id: %v, err: %v", CursorId, err)
	}
	if cursor == nil {
		cursor = gql.NewCursorInstance(CursorId, "")
	}
	return cursor, nil
}

func (m *Doccache) mutate(mutation *gql.Mutation, cursor string) error {
	m.Cursor.SetValue("cursor", cursor)
	cursorMutation := m.Cursor.AddMutation(true)
	return m.client.Mutate(mutation, cursorMutation)
}

func (m *Doccache) UpdateCursor(cursor string) error {
	m.Cursor.Values["cursor"] = cursor
	err := m.client.Mutate(m.Cursor.AddMutation(true))
	if err != nil {
		return fmt.Errorf("failed to update cursor, value: %v, error: %v", cursor, err)
	}
	return nil
}

func (m *Doccache) updateSchemaType(simplifiedType *gql.SimplifiedType) (gql.SchemaUpdateOp, error) {
	updateOp, err := m.Schema.UpdateType(simplifiedType)
	if err != nil {
		return gql.SchemaUpdateOp_None, fmt.Errorf("failed updating local schema, error: %v", err)
	}
	if updateOp != gql.SchemaUpdateOp_None {
		err := m.admin.UpdateSchema(m.Schema)
		if err != nil {
			return gql.SchemaUpdateOp_None, fmt.Errorf("failed updating remote schema, error: %v", err)
		}
	}
	return updateOp, nil
}

func (m *Doccache) addSchemaEdge(typeName, edgeName, edgeType string) error {
	added, err := m.Schema.AddEdge(typeName, edgeName, edgeType)
	if err != nil {
		return fmt.Errorf("failed updating local schema, error: %v", err)
	}
	if added {
		err := m.admin.UpdateSchema(m.Schema)
		if err != nil {
			return fmt.Errorf("failed updating remote schema, error: %v", err)
		}
	}
	return nil
}

func (m *Doccache) GetInstance(hash interface{}, simplifiedType *gql.SimplifiedType, projection []string) (*gql.SimplifiedInstance, error) {
	return m.client.GetOne(hash, simplifiedType, projection)
}

func (m *Doccache) GetInstances(hashes []interface{}, simplifiedType *gql.SimplifiedType, projection []string) (map[interface{}]*gql.SimplifiedInstance, error) {
	return m.client.Get(hashes, simplifiedType, projection)
}

//StoreDocument Creates a new document or updates its certificates
func (m *Doccache) StoreDocument(chainDoc *domain.ChainDocument, cursor string) error {
	parsedDoc, err := chainDoc.ToParsedDoc()
	if err != nil {
		return fmt.Errorf("failed to store document with hash: %v, error building instance from chain doc: %v", chainDoc.Hash, err)
	}
	instance := parsedDoc.Instance
	currentSimplifiedType, err := m.Schema.GetSimplifiedType(instance.SimplifiedType.Name)
	if err != nil {
		return fmt.Errorf("failed to store document with hash: %v of type: %v, error getting simplified type from schema: %v", chainDoc.Hash, instance.GetValue("type"), err)
	}
	err = m.AddCoreEdges(parsedDoc, currentSimplifiedType)
	if err != nil {
		return fmt.Errorf("failed to store document with hash: %v of type: %v, error adding core edges: %v", chainDoc.Hash, instance.GetValue("type"), err)
	}
	updateOp, err := m.updateSchemaType(instance.SimplifiedType)
	if err != nil {
		return fmt.Errorf("failed to store document with hash: %v of type: %v, error updating schema: %v", chainDoc.Hash, instance.GetValue("type"), err)
	}
	var oldInstance *gql.SimplifiedInstance
	if updateOp != gql.SchemaUpdateOp_Created {
		oldInstance, err = m.GetInstance(instance.GetValue("hash"), currentSimplifiedType, currentSimplifiedType.GetCoreFields())
		if err != nil {
			return fmt.Errorf("failed to store document with hash: %v of type: %v, error fetching old instance: %v", chainDoc.Hash, instance.GetValue("type"), err)
		}
	}

	if oldInstance == nil {
		log.Infof("Creating document: %v of type: %v", chainDoc.Hash, instance.GetValue("type"))
		err = m.mutate(instance.AddMutation(false), cursor)
		if err != nil {
			return fmt.Errorf("failed to create document with hash: %v of type: %v, error inserting instance: %v", chainDoc.Hash, instance.GetValue("type"), err)
		}
	} else {
		//TODO: handle certificates
		log.Infof("Updating document: %v of type: %v", chainDoc.Hash, instance.GetValue("type"))
		mutation, err := instance.UpdateMutation(oldInstance)
		if err != nil {
			return fmt.Errorf("failed to update document with hash: %v of type: %v, error generating update mutation: %v", chainDoc.Hash, instance.GetValue("type"), err)
		}
		err = m.mutate(mutation, cursor)
		if err != nil {
			return fmt.Errorf("failed to update document with hash: %v of type: %v, error updating instance: %v", chainDoc.Hash, instance.GetValue("type"), err)
		}
	}

	return nil
}

func (m *Doccache) AddCoreEdges(parsedDoc *domain.ParsedDoc, currentType *gql.SimplifiedType) error {
	newInstance := parsedDoc.Instance
	newType := newInstance.SimplifiedType
	var missingFields []string
	if currentType == nil {
		missingFields = parsedDoc.ChecksumFields
	} else {
		for _, checksumField := range parsedDoc.ChecksumFields {
			coreEdgeFieldName := GetCoreEdgeName(checksumField)
			field := currentType.GetField(coreEdgeFieldName)
			if field != nil {
				newType.SetField(coreEdgeFieldName, field)
				newInstance.SetValue(coreEdgeFieldName, GetEdgeValue(newInstance.GetValue(checksumField)))
			} else {
				missingFields = append(missingFields, checksumField)
			}
		}
	}

	if len(missingFields) == 0 {
		return nil
	}
	missingChecksums := make([]interface{}, 0, len(missingFields))
	for _, missingField := range missingFields {
		missingChecksums = append(missingChecksums, newInstance.GetValue(missingField))
	}
	instances, err := m.client.Get(missingChecksums, gql.DocumentSimplifiedType, nil)
	if err != nil {
		return fmt.Errorf("failed getting core edge documents, for document: %v of type: %v, error: %v", newInstance.GetValue("hash"), newInstance.GetValue("type"), err)
	}
	for _, field := range missingFields {
		hash := newInstance.GetValue(field)
		instance := instances[hash]
		if instance == nil {
			return fmt.Errorf("core edge with hash: %v not found", hash)
		}
		coreEdgeFieldName := GetCoreEdgeName(field)
		newType.SetField(coreEdgeFieldName, &gql.SimplifiedField{
			Name: coreEdgeFieldName,
			Type: instance.GetValue("type").(string),
		})
		newInstance.SetValue(coreEdgeFieldName, GetEdgeValue(instance.GetValue("hash")))
	}
	return nil
}

func GetCoreEdgeName(checksumFieldName string) string {
	return fmt.Sprintf("%v_%v", checksumFieldName, CoreEdgeSuffix)
}

func GetEdgeValue(hash interface{}) map[string]interface{} {
	return map[string]interface{}{"hash": hash}
}

//DeleteDocument Deletes a document
func (m *Doccache) DeleteDocument(chainDoc *domain.ChainDocument, cursor string) error {
	parsedDoc, err := chainDoc.ToParsedDoc()
	if err != nil {
		return fmt.Errorf("failed to delete document with hash: %v, error building instance from chain doc: %v", chainDoc.Hash, err)
	}
	instance := parsedDoc.Instance
	log.Infof("Deleting Node: %v of type: %v", chainDoc.Hash, instance.GetValue("type"))
	mutation, err := instance.DeleteMutation()
	if err != nil {
		return fmt.Errorf("failed to delete document with hash: %v of type: %v, error creating delete mutation: %v", chainDoc.Hash, instance.GetValue("type"), err)
	}
	err = m.mutate(mutation, cursor)
	if err != nil {
		return fmt.Errorf("failed to delete document with hash: %v of type: %v, error deleting instance: %v", chainDoc.Hash, instance.GetValue("type"), err)
	}
	return nil
}

//MutateEdge Creates/Deletes an edge
func (m *Doccache) MutateEdge(chainEdge *domain.ChainEdge, deleteOp bool, cursor string) error {
	instances, err := m.GetInstances(
		[]interface{}{chainEdge.From, chainEdge.To},
		gql.DocumentSimplifiedType,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v, From: %v, To: %v], Delete Op: %v, failed getting instances, error: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp, err)
	}

	fromInstance, ok := instances[chainEdge.From]
	if !ok {
		return fmt.Errorf("FROM node of the relationship: [Edge: %v, From: %v, To: %v] does not exist, Delete Op: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp)
	}

	toInstance, ok := instances[chainEdge.To]
	if !ok {
		return fmt.Errorf("TO node of the relationship: [Edge: %v, From: %v, To: %v] does not exist, Delete Op: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp)
	}
	fromTypeName := fromInstance.GetValue("type").(string)
	err = m.addSchemaEdge(fromTypeName, chainEdge.Name, toInstance.GetValue("type").(string))
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v, From: %v, To: %v], Delete Op: %v, failed updating schema, error: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp, err)
	}

	fromType, err := m.Schema.GetSimplifiedType(fromTypeName)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v, From: %v, To: %v], Delete Op: %v, failed getting type: %v, error: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp, fromInstance.SimplifiedType.Name, err)
	}
	var set, remove map[string]interface{}
	if deleteOp {
		remove = chainEdge.GetEdgeRef()
	} else {
		set = chainEdge.GetEdgeRef()
	}
	mutation, err := fromType.UpdateMutation(chainEdge.From, set, remove)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v, From: %v, To: %v], Delete Op: %v, failed creating edge mutation, error: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp, err)
	}
	log.Infof("Mutating [Edge: %v, From: %v, To: %v] Delete Op: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp)
	err = m.mutate(mutation, cursor)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v, From: %v, To: %v], Delete Op: %v, failed storing edge, error: %v", chainEdge.Name, chainEdge.From, chainEdge.To, deleteOp, err)
	}
	return nil
}
