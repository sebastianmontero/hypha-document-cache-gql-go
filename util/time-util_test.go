package util_test

import (
	"fmt"
	"testing"

	"github.com/sebastianmontero/hypha-document-cache-gql-go/util"
)

func TestToTime(t *testing.T) {
	tt := util.ToTime("2020-11-12T18:27:47.000Z")
	fmt.Println("Time1: ", tt)
	tt = util.ToTime("2020-11-12T18:27:47.5Z")
	fmt.Println("Time1: ", tt)
	tt = util.ToTime("2020-11-12T18:27:47Z")
	fmt.Println("Time1: ", tt)
}
