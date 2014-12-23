package main

import (
	"os"
	"strconv"

	"github.com/couchbaselabs/logg"
	et "github.com/tleyden/elastic-thought"
)

func init() {
	et.EnableAllLogKeys()
}

func main() {

	config := *(et.NewDefaultConfiguration())

	numCbfsNodes := os.Args[1]

	if numCbfsNodes == "" {
		logg.LogFatal("Must pass in the number of cbfs nodes as 1st arg")
		return
	}

	numCbfsNodesInt, err := strconv.ParseInt(numCbfsNodes, 10, 64)
	if err != nil {
		logg.LogFatal("Could not parse %v into int", numCbfsNodes)
		return
	}

	config.NumCbfsClusterNodes = int(numCbfsNodesInt)

	if err = et.EnvironmentSanityCheck(config); err != nil {
		logg.LogFatal("Failed environment sanity check: %v", err)
		return
	}

}
