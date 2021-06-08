package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/dfuse-io/bstream"
	pbcodec "github.com/dfuse-io/dfuse-eosio/pb/dfuse/eosio/codec/v1"
	pbbstream "github.com/dfuse-io/pbgo/dfuse/bstream/v1"
	"github.com/rs/zerolog"
	"github.com/sebastianmontero/dfuse-firehose-client/dfclient"
	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-go/doccache"
	"github.com/sebastianmontero/hypha-document-cache-go/monitoring"
	"github.com/sebastianmontero/hypha-document-cache-go/monitoring/metrics"
	"github.com/sebastianmontero/slog-go/slog"
)

var (
	contract  = os.Getenv("CONTRACT_NAME")
	docTable  = os.Getenv("DOC_TABLE_NAME")
	edgeTable = os.Getenv("EDGE_TABLE_NAME")
	log       *slog.Log
)

type deltaStreamHandler struct {
	cursor   string
	doccache *doccache.Doccache
}

func (m *deltaStreamHandler) OnDelta(delta *dfclient.TableDelta, cursor string, forkStep pbbstream.ForkStep) {
	log.Debugf("On Delta: \nCursor: %v \nFork Step: %v \nDelta %v ", cursor, forkStep, delta)
	if delta.TableName == docTable {
		chainDoc := &doccache.ChainDocument{}
		switch delta.Operation {
		case pbcodec.DBOp_OPERATION_INSERT, pbcodec.DBOp_OPERATION_UPDATE:
			err := json.Unmarshal(delta.NewData, chainDoc)
			if err != nil {
				log.Panicf(err, "Error unmarshalling doc new data: %v", string(delta.NewData))
			}
			log.Tracef("Storing doc: ", chainDoc)
			err = m.doccache.StoreDocument(chainDoc, cursor)
			if err != nil {
				log.Panicf(err, "Failed to store doc: %", chainDoc)
			}
			metrics.CreatedDocs.Inc()
		case pbcodec.DBOp_OPERATION_REMOVE:
			err := json.Unmarshal(delta.OldData, chainDoc)
			if err != nil {
				log.Panicf(err, "Error unmarshalling doc old data: %v", string(delta.OldData))
			}
			err = m.doccache.DeleteDocument(chainDoc, cursor)
			if err != nil {
				log.Panicf(err, "Failed to delete doc: ", chainDoc)
			}
			metrics.DeletedDocs.Inc()
		}
	} else if delta.TableName == edgeTable {
		switch delta.Operation {
		case pbcodec.DBOp_OPERATION_INSERT, pbcodec.DBOp_OPERATION_REMOVE:
			var (
				deltaData []byte
				deleteOp  bool
			)
			chainEdge := &doccache.ChainEdge{}
			if delta.Operation == pbcodec.DBOp_OPERATION_INSERT {
				deltaData = delta.NewData
			} else {
				deltaData = delta.OldData
				deleteOp = true
			}
			err := json.Unmarshal(deltaData, chainEdge)
			if err != nil {
				log.Panicf(err, "Error unmarshalling edge data: ", chainEdge)
			}
			err = m.doccache.MutateEdge(chainEdge, deleteOp, cursor)
			if err != nil {
				log.Errorf(err, "Failed to mutate doc, deleteOp: %v, edge: %v", deleteOp, chainEdge)
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
		log.Panicf(err, "Failed to update cursor: ", cursor)
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
	firehoseEndpoint := os.Getenv("FIREHOSE_ENDPOINT")
	dfuseAPIKey := os.Getenv("DFUSE_API_KEY")
	eosEndpoint := os.Getenv("EOS_ENDPOINT")
	dgraphEndpoint := fmt.Sprintf("%v:%v", os.Getenv("DGRAPH_ALPHA_HOST"), os.Getenv("DGRAPH_ALPHA_EXTERNAL_PORT"))

	startBlock, err := strconv.ParseInt(os.Getenv("START_BLOCK"), 10, 64)
	if err != nil {
		log.Panicf(err, "Unable to parse start block: %v", os.Getenv("START_BLOCK"))
	}

	prometheusPort, err := strconv.ParseUint(os.Getenv("PROMETHEUS_PORT"), 10, 32)
	if err != nil {
		log.Panicf(err, "Unable to parse prometheus port: %v", os.Getenv("PROMETHEUS_PORT"))
	}

	heartBeatFrequency, err := strconv.ParseUint(os.Getenv("HEART_BEAT_FREQUENCY"), 10, 32)
	if err != nil {
		log.Panicf(err, "Unable to parse prometheus port: %v", os.Getenv("PROMETHEUS_PORT"))
	}

	log.Infof(
		`Env Vars
		 contract: %v
		 docTable: %v
		 edgeTable: %v
		 firehoseEndpoint: %v
		 eosEndpoint: %v
		 dgraphEndpoint: %v
		 startBlock: %v
		 prometheusPort: %v
		 heartBeatFrequency: %v`,
		contract,
		docTable,
		edgeTable,
		firehoseEndpoint,
		eosEndpoint,
		dgraphEndpoint,
		startBlock,
		prometheusPort,
		heartBeatFrequency,
	)

	go monitoring.SetupEndpoint(uint(prometheusPort))
	if err != nil {
		log.Panic(err, "Error seting up prometheus endpoint")
	}

	client, err := dfclient.NewDfClient(firehoseEndpoint, dfuseAPIKey, eosEndpoint, nil)
	if err != nil {
		log.Panic(err, "Error creating dfclient")
	}
	dg, err := dgraph.New(dgraphEndpoint)
	if err != nil {
		log.Panic(err, "Error creating dgraph client")
	}
	cache, err := doccache.New(dg, nil)
	if err != nil {
		log.Panic(err, "Error creating doccache client")
	}
	log.Infof("Cursor: %v", cache.Cursor)
	deltaRequest := &dfclient.DeltaStreamRequest{
		StartBlockNum:      startBlock,
		StartCursor:        cache.Cursor.Cursor,
		StopBlockNum:       0,
		ForkSteps:          []pbbstream.ForkStep{pbbstream.ForkStep_STEP_NEW, pbbstream.ForkStep_STEP_UNDO},
		ReverseUndoOps:     true,
		HeartBeatFrequency: uint(heartBeatFrequency),
	}
	// deltaRequest.AddTables("eosio.token", []string{"balance"})
	deltaRequest.AddTables(contract, []string{docTable, edgeTable})
	client.DeltaStream(deltaRequest, &deltaStreamHandler{
		doccache: cache,
	})
}
