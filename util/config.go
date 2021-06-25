package util

import (
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
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
			return nil, err
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
		typeName := mapping["type"].(string)
		if len(fullLabels) == 0 {
			return nil, fmt.Errorf("type mapping for type: %v has no labels", typeName)
		}
		typeMappings[typeName] = fullLabels
	}
	return typeMappings, nil
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
