# Document Cache Stream Processor

Connects to the dfuse firehose to get a stream of table deltas from a contract that implements document graph, and updates the graphql schema and the dgraph database accordingly, in order for the dgraph database to reflect the data contained in the documents and edges contract tables.

To run only the document cache process with the helper script:

`./start-local-env.sh`

Or:

`go run . ./config.yml`

Look at the **config.yml** file for an example configuration file, some important configuration parameters are:

- firehose-endpoint: The dfuse firehose endpoint to connecto
- dgraph-*: Specifies the parameters to connect to the dgraph services
- type-mappings: Provides the details to enable the document cache process to determine the type of a document based on its properties
- custom-interfaces: Defines interfaces to be created and the types that should implement them
- logical-ids: Defines additional ids for types

An additional convinience script is provided to run both dgraph and the document cache process as docker containers:

To run dgraph and document cache by building the document cache image based on the current state of the project for testnet:

`./start-doc-cache.sh testnet build`

To run dgraph and document cache by pulling the latest document cache image from docker hub for mainnet:

`./start-doc-cache.sh mainnet image`

When using the **start-doc-cache.sh** script the document cache process is configured using environment variables, look at .env.mainnet and .env.testnet for examples.

