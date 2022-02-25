package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dfuse-io/bstream"
	pbcodec "github.com/dfuse-io/dfuse-eosio/pb/dfuse/eosio/codec/v1"
	pbbstream "github.com/dfuse-io/pbgo/dfuse/bstream/v1"
	"github.com/rs/zerolog"
	"github.com/sebastianmontero/dfuse-firehose-client/dfclient"
	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/config"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/doccache/domain"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/monitoring"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/monitoring/metrics"
	"github.com/sebastianmontero/slog-go/slog"
)

var (
	log *slog.Log
)

type deltaStreamHandler struct {
	cursor   string
	doccache *doccache.Doccache
	config   *config.Config
}

func (m *deltaStreamHandler) OnDelta(delta *dfclient.TableDelta, cursor string, forkStep pbbstream.ForkStep) {
	log.Debugf("On Delta: \nCursor: %v \nFork Step: %v \nDelta %v ", cursor, forkStep, delta)
	log.Debugf("Doc table name: %v ", m.config.DocTableName)
	if delta.TableName == m.config.DocTableName {
		chainDoc := &domain.ChainDocument{}
		switch delta.Operation {
		case pbcodec.DBOp_OPERATION_INSERT, pbcodec.DBOp_OPERATION_UPDATE:
			err := json.Unmarshal(delta.NewData, chainDoc)
			if err != nil {
				log.Panicf(err, "Error unmarshalling doc new data: %v", string(delta.NewData))
			}
			log.Tracef("Storing doc: %v ", chainDoc)
			err = m.doccache.StoreDocument(chainDoc, cursor)
			if err != nil {
				log.Panicf(err, "Failed to store doc: %v", chainDoc)
			}
			metrics.CreatedDocs.Inc()
		case pbcodec.DBOp_OPERATION_REMOVE:
			err := json.Unmarshal(delta.OldData, chainDoc)
			if err != nil {
				log.Panicf(err, "Error unmarshalling doc old data: %v", string(delta.OldData))
			}
			err = m.doccache.DeleteDocument(chainDoc, cursor)
			if err != nil {
				log.Panicf(err, "Failed to delete doc: %v", chainDoc)
			}
			metrics.DeletedDocs.Inc()
		}
	} else if delta.TableName == m.config.EdgeTableName {
		switch delta.Operation {
		case pbcodec.DBOp_OPERATION_INSERT, pbcodec.DBOp_OPERATION_REMOVE:
			var (
				deltaData []byte
				deleteOp  bool
			)
			chainEdge := &domain.ChainEdge{}
			if delta.Operation == pbcodec.DBOp_OPERATION_INSERT {
				deltaData = delta.NewData
			} else {
				deltaData = delta.OldData
				deleteOp = true
			}
			err := json.Unmarshal(deltaData, chainEdge)
			if err != nil {
				log.Panicf(err, "Error unmarshalling edge data: %v", chainEdge)
			}
			err = m.doccache.MutateEdge(chainEdge, deleteOp, cursor)
			if err != nil {
				log.Panicf(err, "Failed to mutate doc, deleteOp: %v, edge: %v", deleteOp, chainEdge)
			}
			if deleteOp {
				metrics.DeletedEdges.Inc()
			} else {
				metrics.CreatedEdges.Inc()
			}

		case pbcodec.DBOp_OPERATION_UPDATE:
			log.Panicf(nil, "Edge updating is not handled: %v", delta)
		}
	}
	metrics.BlockNumber.Set(float64(delta.Block.Number))
	m.cursor = cursor
}

func (m *deltaStreamHandler) OnHeartBeat(block *pbcodec.Block, cursor string) {
	err := m.doccache.UpdateCursor(cursor)
	if err != nil {
		log.Panicf(err, "Failed to update cursor: %v", cursor)
	}
	metrics.BlockNumber.Set(float64(block.Number))
}

func (m *deltaStreamHandler) OnError(err error) {
	log.Error(err, "On Error")
}

func (m *deltaStreamHandler) OnComplete(lastBlockRef bstream.BlockRef) {
	log.Infof("On Complete Last Block Ref: %v", lastBlockRef)
}

func main() {
	log = slog.New(&slog.Config{Pretty: true, Level: zerolog.DebugLevel}, "start-doccache")
	if len(os.Args) != 2 {
		log.Panic(nil, "Config file has to be specified as the only cmd argument")
	}
	log.Infof("Sleeping for a while to let dgraph start...")
	time.Sleep(time.Minute * 4)
	config, err := config.LoadConfig(os.Args[1])
	if err != nil {
		log.Panicf(err, "Unable to load config file: %v", os.Args[1])
	}

	log.Info(config.String())

	go monitoring.SetupEndpoint(config.PrometheusPort)
	if err != nil {
		log.Panic(err, "Error seting up prometheus endpoint")
	}

	client, err := dfclient.NewDfClient(config.FirehoseEndpoint, config.DfuseApiKey, config.EosEndpoint, nil)
	if err != nil {
		log.Panic(err, "Error creating dfclient")
	}
	dg, err := dgraph.New(config.DgraphGRPCEndpoint)
	if err != nil {
		log.Panic(err, "Error creating dgraph client")
	}
	gqlAdmin := gql.NewAdmin(config.GQLAdminURL)
	gqlClient := gql.NewClient(config.GQLClientURL)
	cache, err := doccache.New(dg, gqlAdmin, gqlClient, config, nil)
	if err != nil {
		log.Panic(err, "Error creating doccache client")
	}
	log.Infof("Cursor: %v", cache.Cursor)
	deltaRequest := &dfclient.DeltaStreamRequest{
		StartBlockNum:      config.StartBlock,
		StartCursor:        cache.Cursor.GetValue("cursor").(string),
		StopBlockNum:       0,
		ForkSteps:          []pbbstream.ForkStep{pbbstream.ForkStep_STEP_NEW, pbbstream.ForkStep_STEP_UNDO},
		ReverseUndoOps:     true,
		HeartBeatFrequency: config.HeartBeatFrequency,
	}
	// deltaRequest.AddTables("eosio.token", []string{"balance"})
	deltaRequest.AddTables(config.ContractName, []string{config.DocTableName, config.EdgeTableName})
	client.DeltaStream(deltaRequest, &deltaStreamHandler{
		doccache: cache,
		config:   config,
	})
}
