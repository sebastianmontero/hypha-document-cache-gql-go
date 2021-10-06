package config

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/spf13/viper"
)

type Config struct {
	ContractName        string                   `mapstructure:"contract-name"`
	DocTableName        string                   `mapstructure:"doc-table-name"`
	EdgeTableName       string                   `mapstructure:"edge-table-name"`
	FirehoseEndpoint    string                   `mapstructure:"firehose-endpoint"`
	EosEndpoint         string                   `mapstructure:"eos-endpoint"`
	DgraphAlphaHost     string                   `mapstructure:"dgraph-alpha-host"`
	DgraphAlphaGRPCPort uint                     `mapstructure:"dgraph-alpha-grpc-port"`
	DgraphAlphaHTTPPort uint                     `mapstructure:"dgraph-alpha-http-port"`
	PrometheusPort      uint                     `mapstructure:"prometheus-port"`
	StartBlock          int64                    `mapstructure:"start-block"`
	HeartBeatFrequency  uint                     `mapstructure:"heart-beat-frequency"`
	DfuseApiKey         string                   `mapstructure:"dfuse-api-key"`
	TypeMappingsRaw     []map[string]interface{} `mapstructure:"type-mappings"`
	TypeMappings        map[string][]string
	InterfacesRaw       []map[string]interface{} `mapstructure:"custom-interfaces"`
	Interfaces          gql.SimplifiedInterfaces
	LogicalIdsRaw       []map[string]interface{} `mapstructure:"logical-ids"`
	LogicalIds          domain.LogicalIds
	DgraphGRPCEndpoint  string
	DgraphHTTPURL       string
	GQLAdminURL         string
	GQLClientURL        string
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(filePath string) (*Config, error) {
	viper.SetConfigFile(filePath)

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	config.DgraphGRPCEndpoint = fmt.Sprintf("%v:%v", config.DgraphAlphaHost, config.DgraphAlphaGRPCPort)
	config.DgraphHTTPURL = fmt.Sprintf("http://%v:%v", config.DgraphAlphaHost, config.DgraphAlphaHTTPPort)
	config.GQLAdminURL = joinUrl(config.DgraphHTTPURL, "admin")
	config.GQLClientURL = joinUrl(config.DgraphHTTPURL, "graphql")
	if config.TypeMappingsRaw != nil {
		config.TypeMappings, err = processTypeMappings(config.TypeMappingsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse type mappings configuration, error: %v", err)
		}
	}
	if config.InterfacesRaw != nil {
		config.Interfaces, err = parseInterfaceConfig(config.InterfacesRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse interfaces configuration, error: %v", err)
		}
	}
	if config.LogicalIdsRaw != nil {
		config.LogicalIds, err = parseLogicalIdsConfig(config.LogicalIdsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse logical ids configuration, error: %v", err)
		}
	}
	return &config, nil
}

func processTypeMappings(raw []map[string]interface{}) (map[string][]string, error) {
	typeMappings := make(map[string][]string)
	for _, mapping := range raw {
		fullLabels := make([]string, 0)
		for groupLabel, labels := range mapping["labels"].(map[interface{}]interface{}) {
			for _, label := range labels.([]interface{}) {
				fullLabels = append(fullLabels, fmt.Sprintf("%v_%v", groupLabel, strcase.ToLowerCamel(label.(string))))
			}
		}
		typeName := domain.GetObjectTypeName(mapping["type"].(string))
		if len(fullLabels) == 0 {
			return nil, fmt.Errorf("type mapping for type: %v has no labels", typeName)
		}
		typeMappings[typeName] = fullLabels
	}
	return typeMappings, nil
}

func parseInterfaceConfig(config []map[string]interface{}) (gql.SimplifiedInterfaces, error) {
	interfaces := gql.NewSimplifiedInterfaces()
	for _, interfConfig := range config {
		name := interfConfig["name"].(string)
		fields := make(map[string]*gql.SimplifiedField)
		signatureFields := make([]string, 0)
		for _, fieldConfigI := range interfConfig["fields"].([]interface{}) {

			fieldConfig := fieldConfigI.(map[interface{}]interface{})
			fieldContentGroup, hasContentGroup := fieldConfig["content-group"].(string)
			fieldName := fieldConfig["name"].(string)
			fieldType := fieldConfig["type"].(string)
			isID, _ := fieldConfig["is-id"].(bool)
			// nonNull, _ := fieldConfig["non-null"].(bool)
			signature, _ := fieldConfig["signature"].(bool)
			var fullFieldName string
			var gqlType string
			var index string
			isArray := false

			if isID && !domain.IsIDableType(fieldType) {
				return nil, fmt.Errorf("id fields can only be of IDable types(checksum, name, string), found type: %v for field: %v of interface: %v", fieldType, fieldName, name)
			}

			if hasContentGroup {
				prefix := domain.GetFieldPrefix(fieldContentGroup)
				if !domain.IsBaseType(fieldType) {
					//Assume base type is checksum pointing to an object of this type
					fullFieldName = domain.GetFieldName(
						prefix,
						fieldName,
						domain.ContentType_Checksum256,
					)
					edgeFieldName := domain.GetCoreEdgeName(fullFieldName)
					fields[edgeFieldName] = &gql.SimplifiedField{
						Name: edgeFieldName,
						Type: domain.GetObjectTypeName(fieldType),
					}
					fieldType = domain.ContentType_Checksum256
				} else {
					fullFieldName = domain.GetFieldName(
						prefix,
						fieldName,
						fieldType,
					)
				}
				gqlType = domain.GetGQLType(fieldType)
				index = domain.GetIndex(fieldType)
			} else {
				if domain.IsBaseType(fieldType) {
					fullFieldName = strcase.ToLowerCamel(fieldName)
					gqlType = domain.GetGQLType(fieldType)
					index = domain.GetIndex(fieldType)
				} else {
					fullFieldName = fieldName
					gqlType = domain.GetObjectTypeName(fieldType)
					isArray = true
				}
			}
			fields[fullFieldName] = &gql.SimplifiedField{
				IsID:    isID,
				Name:    fullFieldName,
				Type:    gqlType,
				Index:   index,
				IsArray: isArray,
				NonNull: isID,
			}
			if signature {
				signatureFields = append(signatureFields, fullFieldName)
			}
		}
		var types []string
		typesI, ok := interfConfig["types"].([]interface{})
		if ok {
			types = make([]string, 0, len(typesI))
			for _, t := range typesI {
				types = append(types, domain.GetObjectTypeName(t.(string)))
			}
		}
		interf := gql.NewSimplifiedInterface(name, fields, signatureFields, types)
		err := interf.Validate()
		if err != nil {
			return nil, err
		}
		interfaces.Put(interf)
	}
	return interfaces, nil
}

func parseLogicalIdsConfig(config []map[string]interface{}) (domain.LogicalIds, error) {
	logicalIds := domain.NewLogicalIds()
	for _, typeConfig := range config {
		objType := domain.GetObjectTypeName(typeConfig["type"].(string))
		idsConfig := typeConfig["ids"].([]interface{})
		ids := make([]string, 0, len(idsConfig))
		for _, idConfigI := range idsConfig {

			idConfig := idConfigI.(map[interface{}]interface{})
			idContentGroup := idConfig["content-group"].(string)
			idName := idConfig["name"].(string)
			idType := idConfig["type"].(string)

			if !domain.IsIDableType(idType) {
				return nil, fmt.Errorf("id fields can only be of IDable types(checksum, name, string), found type: %v for field: %v of object: %v", idType, idName, objType)
			}
			fullIdName := domain.GetFieldName(
				domain.GetFieldPrefix(idContentGroup),
				idName,
				idType,
			)
			ids = append(ids, fullIdName)
		}
		logicalIds.Set(objType, ids)
	}
	return logicalIds, nil
}

func (m *Config) String() string {
	return fmt.Sprintf(
		`
			Config {
				ContractName: %v
				DocTableName: %v
				EdgeTableName: %v
				FirehoseEndpoint: %v
				EosEndpoint: %v
				DgraphAlphaHost: %v
				DgraphAlphaGRPCPort: %v
				DgraphAlphaHTTPPort: %v
				DgraphGRPCEndpoint: %v
				DgraphHTTPURL: %v      
				PrometheusPort: %v
				StartBlock: %v
				HeartBeatFrequency: %v
				DfuseApiKey: %v
				TypeMappingsRaw: %v
				TypeMappings: %v
				GQLAdminURL: %v
				GQLClientURL: %v
			}
		`,
		m.ContractName,
		m.DocTableName,
		m.EdgeTableName,
		m.FirehoseEndpoint,
		m.EosEndpoint,
		m.DgraphAlphaHost,
		m.DgraphAlphaGRPCPort,
		m.DgraphAlphaHTTPPort,
		m.DgraphGRPCEndpoint,
		m.DgraphHTTPURL,
		m.PrometheusPort,
		m.StartBlock,
		m.HeartBeatFrequency,
		m.DfuseApiKey,
		m.TypeMappingsRaw,
		m.TypeMappings,
		m.GQLAdminURL,
		m.GQLClientURL,
	)
}

func joinUrl(base, path string) string {
	return fmt.Sprintf("%v/%v", strings.TrimRight(base, "/"), path)
}
