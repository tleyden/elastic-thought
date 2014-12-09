package main

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type Params struct {
	COUCHBASE_SERVER bool
	CBFS             bool
	SYNC_GATEWAY     bool
	ELASTIC_THOUGHT  bool
	GPU              bool
}

func main() {

	if len(os.Args) < 3 {
		log.Fatal("Usage: ./generate_cloudformation elastic_thought (cpu|gpu)")
		return
	}

	params := Params{}

	switch os.Args[1] {
	case "elastic_thought":
		params.COUCHBASE_SERVER = true
		params.CBFS = true
		params.SYNC_GATEWAY = true
		params.ELASTIC_THOUGHT = true
	case "sync_gateway":
		params.COUCHBASE_SERVER = true
		params.CBFS = true
		params.SYNC_GATEWAY = true
		params.ELASTIC_THOUGHT = false
	case "cbfs":
		params.COUCHBASE_SERVER = true
		params.CBFS = true
		params.SYNC_GATEWAY = false
		params.ELASTIC_THOUGHT = false
	case "couchbase_server":
		params.COUCHBASE_SERVER = true
		params.CBFS = false
		params.SYNC_GATEWAY = false
		params.ELASTIC_THOUGHT = false
	default:
		log.Fatal("invalid arg: %v", os.Args[1])
	}

	switch os.Args[2] {
	case "cpu":
		params.GPU = false
	case "gpu":
		params.GPU = true
	default:
		log.Fatal("invalid arg: %v", os.Args[2])
	}

	templateFile := "cloudformation.template"

	templateBytes, err := ioutil.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("cloudformation").Parse(string(templateBytes))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, params)
	if err != nil {
		panic(err)
	}

}
