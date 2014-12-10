package main

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type Params struct {
	ProcessorType string // cpu or gpu
	CaffeBranch   string // develop or master
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: ./generate_dockerfiles (cpu|gpu) (develop|master)")
		return
	}

	params := Params{}

	switch os.Args[1] {
	case "cpu":
		params.ProcessorType = os.Args[1]
	case "gpu":
		params.ProcessorType = os.Args[1]
	default:
		log.Fatalf("Invalid argument for cpu|gpu: %v", os.Args[1])
	}

	switch os.Args[2] {
	case "develop":
		params.CaffeBranch = os.Args[2]
	case "master":
		params.CaffeBranch = os.Args[2]
	default:
		log.Fatalf("Invalid argument for develop|master: %v", os.Args[2])
	}

	templateFile := "../templates/Dockerfile.template"

	templateBytes, err := ioutil.ReadFile(templateFile)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.New("docker").Parse(string(templateBytes))
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, params)
	if err != nil {
		panic(err)
	}

}
