module github.com/sebastianmontero/hypha-document-cache-gql-go

go 1.15

require (
	github.com/dfuse-io/dfuse-eosio v0.9.0-beta9.0.20210812014530-dcb01c5c4b35
	github.com/iancoleman/strcase v0.1.3
	github.com/machinebox/graphql v0.2.2
	github.com/matryer/is v1.4.0 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/rs/zerolog v1.20.0
	github.com/sebastianmontero/dfuse-firehose-client v0.0.0-20220927181845-114d1d801c5d
	github.com/sebastianmontero/dgraph-go-client v0.0.0-20210213215931-344d1e456654
	github.com/sebastianmontero/slog-go v0.0.0-20210213204103-60eda76e8d74
	github.com/spf13/viper v1.8.0
	github.com/streamingfast/bstream v0.0.2-0.20210811181043-4c1920a7e3e3
	github.com/streamingfast/pbgo v0.0.6-0.20210811160400-7c146c2db8cc
	github.com/vektah/gqlparser v1.3.1
	gotest.tools v2.2.0+incompatible
)

// replace github.com/sebastianmontero/dfuse-firehose-client => ../dfuse-firehose-client
