module github.com/sebastianmontero/hypha-document-cache-gql-go

go 1.15

require (
	github.com/iancoleman/strcase v0.2.0
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/pinax-network/firehose-antelope v1.0.2
	github.com/prometheus/client_golang v1.16.0
	github.com/rs/zerolog v1.28.0
	github.com/sebastianmontero/dfuse-firehose-client v0.0.0-20240116211353-9c05b81d7583
	github.com/sebastianmontero/dgraph-go-client v0.0.0-20210213215931-344d1e456654
	github.com/sebastianmontero/slog-go v0.0.0-20210213204103-60eda76e8d74
	github.com/spf13/viper v1.15.0
	github.com/streamingfast/pbgo v0.0.6-0.20231208140754-ed2bd10b96ee
	github.com/vektah/gqlparser v1.3.1
	gotest.tools v2.2.0+incompatible
)

// replace github.com/sebastianmontero/dfuse-firehose-client => ../dfuse-firehose-client
