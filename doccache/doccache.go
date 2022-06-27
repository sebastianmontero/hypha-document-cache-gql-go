package doccache

import (
	"fmt"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/config"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/slog-go/slog"
)

const CursorIdName string = "id"
const CursorIdValue string = "c1"
const DoccacheConfigIdValue string = "dc1"
const DocumentIdName string = "docId"

var log *slog.Log

//Doccache Service class to store and retrieve docs from dgraph
type Doccache struct {
	dgraph *dgraph.Dgraph
	admin  *gql.Admin
	client *gql.Client
	config *config.Config
	Cursor *gql.SimplifiedInstance
	Schema *gql.Schema
}

//New creates a new doccache instance
func New(dg *dgraph.Dgraph, admin *gql.Admin, client *gql.Client, config *config.Config, logConfig *slog.Config) (*Doccache, error) {
	log = slog.New(logConfig, "doccache")

	m := &Doccache{
		dgraph: dg,
		admin:  admin,
		client: client,
		config: config,
	}

	err := m.PrepareSchema()
	if err != nil {
		return nil, err
	}
	err = m.updateDoccacheConfig()
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

// Sets up the base gql schema to be used based on the initial configuration
func (m *Doccache) PrepareSchema() error {
	log.Infof("Getting current schema...")
	schema, err := m.admin.GetCurrentSchema()
	fmt.Println("Current schema: ", schema)
	if err != nil {
		return fmt.Errorf("failed getting current schema, error: %v", err)
	}
	if schema == nil {
		log.Infof("No current schema, creating initial schema...")
		schema, err = gql.InitialSchema()
		if err != nil {
			return fmt.Errorf("failed getting initial schema, error: %v", err)
		}
		err = m.admin.UpdateSchema(schema)
		if err != nil {
			return fmt.Errorf("failed setting initial schema, error: %v", err)
		}
		log.Infof("Created initial schema.")
	}
	err = m.initializeInterfacesSchema(schema)
	if err != nil {
		return fmt.Errorf("failed initializing interfaces schema error: %v", err)
	}
	m.Schema = schema
	return nil
}

// Sets up the gql schema for the interfaces specified in the initial configuration
func (m *Doccache) initializeInterfacesSchema(schema *gql.Schema) error {
	log.Infof("Initializing interfaces schema...")
	for _, simplifiedInterface := range m.config.Interfaces {
		interf := schema.GetType(simplifiedInterface.Name)
		if interf == nil {
			log.Infof("Interface: %v not found creating...", simplifiedInterface.Name)
			objFields := m.config.Interfaces.GetObjectTypeFields(simplifiedInterface.Name)
			for _, objField := range objFields {
				obj := schema.GetType(objField.Type)
				if obj == nil {
					log.Infof("Object type: '%v' of field: '%v' part of interface: '%v' not found creating...", objField.Type, objField.Name, simplifiedInterface.Name)
					_, err := schema.UpdateType(gql.NewSimplifiedType(objField.Type, nil, gql.DocumentSimplifiedInterface))
					if err != nil {
						return fmt.Errorf("failed adding type: %v for field: %v of interface: %v, error: %v", objField.Type, objField.Name, simplifiedInterface.Name, err)
					}
				}
			}
			schema.SetInterface(simplifiedInterface)
		}
		err := m.admin.UpdateSchema(schema)
		if err != nil {
			return fmt.Errorf("failed initializing interfaces schema, error: %v", err)
		}
	}
	return nil
}

// Finds the current cursor
func (m *Doccache) getCursor() (*gql.SimplifiedInstance, error) {

	cursor, err := m.client.GetOne(CursorIdName, CursorIdValue, gql.CursorSimplifiedType, nil)
	if err != nil {
		return nil, fmt.Errorf("failed getting cursor with id: %v, err: %v", CursorIdValue, err)
	}
	if cursor == nil {
		cursor = gql.NewCursorInstance(CursorIdValue, "")
	}
	return cursor, nil
}

// Creates/Updates the docache configuration object stored on the db
func (m *Doccache) updateDoccacheConfig() error {

	_, err := m.updateSchemaType(gql.DoccacheConfigSimplifiedType)
	if err != nil {
		return fmt.Errorf("failed to update schema with the doccache config type, type: %v, error : %v", gql.DoccacheConfigSimplifiedType, err)
	}

	doccacheConfig := gql.NewSimplifiedInstance(
		gql.DoccacheConfigSimplifiedType,
		map[string]interface{}{
			"id":              DoccacheConfigIdValue,
			"contract":        m.config.ContractName,
			"eosEndpoint":     m.config.EosEndpoint,
			"documentsTable":  m.config.DocTableName,
			"edgesTable":      m.config.EdgeTableName,
			"elasticEndpoint": m.config.ElasticEndpoint,
			"elasticApiKey":   m.config.ElasticApiKey,
		},
	)
	err = m.client.Mutate(doccacheConfig.AddMutation(true))
	if err != nil {
		return fmt.Errorf("failed to update doccache config, value: %v, error: %v", doccacheConfig, err)
	}
	return nil
}

// Executes a graphql mutation
func (m *Doccache) mutate(mutation *gql.Mutation, cursor string) error {
	m.Cursor.SetValue("cursor", cursor)
	cursorMutation := m.Cursor.AddMutation(true)
	return m.client.Mutate(mutation, cursorMutation)
}

// Updates the cursor stored on the db
func (m *Doccache) UpdateCursor(cursor string) error {
	m.Cursor.Values["cursor"] = cursor
	err := m.client.Mutate(m.Cursor.AddMutation(true))
	if err != nil {
		return fmt.Errorf("failed to update cursor, value: %v, error: %v", cursor, err)
	}
	return nil
}

// Updates the gql schema for a type based on the differences between the current schema and
// the newly found object found on chain
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

// Updates the schema for an edge based on the newly found edge found on chain
func (m *Doccache) updateSchemaEdge(typeName, edgeName, edgeType string) error {
	added, err := m.Schema.UpdateEdge(typeName, edgeName, edgeType)
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

// Returns the document with the specified id
func (m *Doccache) GetDocumentInstance(docId interface{}, simplifiedType *gql.SimplifiedType, projection []string) (*gql.SimplifiedInstance, error) {
	return m.client.GetOne(DocumentIdName, docId, simplifiedType, projection)
}

// Returns the current cursor
func (m *Doccache) GetCursorInstance(cursorId interface{}, simplifiedType *gql.SimplifiedType, projection []string) (*gql.SimplifiedInstance, error) {
	return m.client.GetOne(CursorIdName, cursorId, simplifiedType, projection)
}

// Returns the current doccache configuration
func (m *Doccache) GetDoccacheConfigInstance() (*gql.SimplifiedInstance, error) {
	return m.client.GetOne("id", DoccacheConfigIdValue, gql.DoccacheConfigSimplifiedType, nil)
}

// Returns the documents with the specified ids, the projection parameter can be used to indicate the properties
// of the document to be returned
func (m *Doccache) GetDocumentBaseInstances(ids []interface{}, simplifiedType *gql.SimplifiedBaseType, projection []string) (map[interface{}]*gql.SimplifiedBaseInstance, error) {
	return m.client.GetBaseInstances(DocumentIdName, ids, simplifiedType, projection)
}

//StoreDocument Creates or updates document
func (m *Doccache) StoreDocument(chainDoc *domain.ChainDocument, cursor string) error {
	parsedDoc, err := chainDoc.ToParsedDoc(m.config.TypeMappings)
	if err != nil {
		return fmt.Errorf("failed to store document with docId: %v, error building instance from chain doc: %v", chainDoc.ID, err)
	}
	instance := parsedDoc.Instance
	newSimplifiedType := instance.SimplifiedType
	currentSimplifiedType, err := m.Schema.GetSimplifiedType(newSimplifiedType.Name)
	if err != nil {
		return fmt.Errorf("failed to store document with docId: %v of type: %v, error getting simplified type from schema: %v", chainDoc.ID, instance.GetValue("type"), err)
	}

	err = m.config.LogicalIds.ConfigureLogicalIds(newSimplifiedType.SimplifiedBaseType)
	if err != nil {
		return fmt.Errorf("failed to store document with docId: %v of type: %v, unable to configure logical ids, error: %v", chainDoc.ID, instance.GetValue("type"), err)
	}
	err = m.config.Interfaces.ApplyInterfaces(newSimplifiedType, currentSimplifiedType)
	if err != nil {
		return fmt.Errorf("failed to store document with docId: %v of type: %v, unable to apply interfaces, error: %v", chainDoc.ID, instance.GetValue("type"), err)
	}

	updateOp, err := m.updateSchemaType(newSimplifiedType)
	if err != nil {
		return fmt.Errorf("failed to store document with docId: %v of type: %v, error updating schema: %v", chainDoc.ID, instance.GetValue("type"), err)
	}
	var oldInstance *gql.SimplifiedInstance
	if updateOp != gql.SchemaUpdateOp_Created {
		oldInstance, err = m.GetDocumentInstance(instance.GetValue(DocumentIdName), currentSimplifiedType, currentSimplifiedType.GetCoreFields())
		if err != nil {
			return fmt.Errorf("failed to store document with docId: %v of type: %v, error fetching old instance: %v", chainDoc.ID, instance.GetValue("type"), err)
		}
	}

	if oldInstance == nil {
		log.Infof("Creating document: %v of type: %v", chainDoc.ID, instance.GetValue("type"))
		err = m.mutate(instance.AddMutation(false), cursor)
		if err != nil {
			return fmt.Errorf("failed to create document with docId: %v of type: %v, error inserting instance: %v", chainDoc.ID, instance.GetValue("type"), err)
		}
	} else {
		log.Infof("Updating document: %v of type: %v", chainDoc.ID, instance.GetValue("type"))
		mutation, err := instance.UpdateMutation(DocumentIdName, oldInstance)
		fmt.Println("Update mutation: ", mutation)
		if err != nil {
			return fmt.Errorf("failed to update document with docId: %v of type: %v, error generating update mutation: %v", chainDoc.ID, instance.GetValue("type"), err)
		}
		err = m.mutate(mutation, cursor)
		if err != nil {
			return fmt.Errorf("failed to update document with docId: %v of type: %v, error updating instance: %v", chainDoc.ID, instance.GetValue("type"), err)
		}
	}

	return nil
}

func GetEdgeValue(docId interface{}) map[string]interface{} {
	return map[string]interface{}{"docId": docId}
}

// Deletes the document represented by the chainDoc parameter
func (m *Doccache) DeleteDocument(chainDoc *domain.ChainDocument, cursor string) error {
	parsedDoc, err := chainDoc.ToParsedDoc(m.config.TypeMappings)
	if err != nil {
		return fmt.Errorf("failed to delete document with docId: %v, error building instance from chain doc: %v", chainDoc.ID, err)
	}
	instance := parsedDoc.Instance
	log.Infof("Deleting Node: %v of type: %v", chainDoc.ID, instance.GetValue("type"))
	mutation, err := instance.DeleteMutation(DocumentIdName)
	if err != nil {
		return fmt.Errorf("failed to delete document with docId: %v of type: %v, error creating delete mutation: %v", chainDoc.ID, instance.GetValue("type"), err)
	}
	err = m.mutate(mutation, cursor)
	if err != nil {
		return fmt.Errorf("failed to delete document with docId: %v of type: %v, error deleting instance: %v", chainDoc.ID, instance.GetValue("type"), err)
	}
	return nil
}

//MutateEdge Creates/Deletes an edge
func (m *Doccache) MutateEdge(chainEdge *domain.ChainEdge, deleteOp bool, cursor string) error {
	instances, err := m.GetDocumentBaseInstances(
		[]interface{}{chainEdge.From, chainEdge.To},
		gql.DocumentSimplifiedInterface.SimplifiedBaseType,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v (%v), From: %v, To: %v], Delete Op: %v, failed getting instances, error: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp, err)
	}

	fromInstance, ok := instances[chainEdge.From]
	if !ok {
		log.Errorf(nil, "FROM node of the relationship: [Edge: %v (%v), From: %v, To: %v] does not exist, Delete Op: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp)
		return nil
	}

	toInstance, ok := instances[chainEdge.To]
	if !ok {
		log.Errorf(nil, "TO node of the relationship: [Edge: %v (%v), From: %v, To: %v] does not exist, Delete Op: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp)
		return nil
	}

	fromTypeName := fromInstance.GetValue("type").(string)
	toTypeName := toInstance.GetValue("type").(string)

	fromType, err := m.Schema.GetSimplifiedType(fromTypeName)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v (%v), From: %v, To: %v], Delete Op: %v, failed getting type: %v, error: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp, fromInstance.SimplifiedBaseType.Name, err)
	}
	edgeType := toTypeName
	currentEdgeField := fromType.GetField(chainEdge.DocEdgeName)
	if currentEdgeField != nil && currentEdgeField.Type != toTypeName {
		edgeType = gql.DocumentSimplifiedInterface.Name
	}
	err = m.updateSchemaEdge(fromTypeName, chainEdge.DocEdgeName, edgeType)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v (%v), From: %v, To: %v], Delete Op: %v, failed updating schema, error: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp, err)
	}

	var set, remove map[string]interface{}
	if deleteOp {
		remove = chainEdge.GetEdgeRef(toInstance.GetValue("docId"))
	} else {
		set = chainEdge.GetEdgeRef(toInstance.GetValue("docId"))
	}
	mutation, err := fromType.UpdateMutation(DocumentIdName, chainEdge.From, set, remove)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v (%v), From: %v, To: %v], Delete Op: %v, failed creating edge mutation, error: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp, err)
	}
	log.Infof("Mutating [Edge: %v (%v), From: %v, To: %v] Delete Op: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp)
	err = m.mutate(mutation, cursor)
	if err != nil {
		return fmt.Errorf("failed mutating edge [Edge: %v (%v), From: %v, To: %v], Delete Op: %v, failed storing edge, error: %v", chainEdge.Name, chainEdge.DocEdgeName, chainEdge.From, chainEdge.To, deleteOp, err)
	}
	return nil
}
