package gql_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/sebastianmontero/dgraph-go-client/dgraph"
	"github.com/sebastianmontero/hypha-document-cache-gql-go/gql"
)

var dg *dgraph.Dgraph
var admin *gql.Admin
var client *gql.Client

func TestMain(m *testing.M) {
	beforeAll()
	// exec test and this returns an exit code to pass to os
	retCode := m.Run()
	afterAll()
	// If exit code is distinct of zero,
	// the test will be failed (red)
	os.Exit(retCode)
}

func beforeAll() {
	admin = gql.NewAdmin("http://localhost:8080/admin")
	client = gql.NewClient("http://localhost:8080/graphql")
	var err error
	dg, err = dgraph.New("")
	if err != nil {
		log.Fatal(err, "Unable to create dgraph")
	}
}

func beforeEach() {
	err := dg.DropAll()
	if err != nil {
		log.Fatal(err, "Unable to drop all")
	}
	time.Sleep(time.Second * 2)
}

func afterAll() {
	dg.Close()
}
