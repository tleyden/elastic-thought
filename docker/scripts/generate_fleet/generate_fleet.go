package main

import (
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

type Params struct {
	ProcessorType string // cpu or gpu
	GPU           bool
	Devices       string
}

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: ./generate_fleet (cpu|gpu)")
		return
	}

	params := Params{}

	switch os.Args[1] {
	case "cpu":
		params.ProcessorType = os.Args[1]
		params.GPU = false
	case "gpu":
		params.ProcessorType = os.Args[1]
		params.GPU = true
	default:
		log.Fatalf("Invalid argument for cpu|gpu: %v", os.Args[1])
	}

	if params.ProcessorType == "gpu" {
		params.Devices = "--device /dev/nvidia0:/dev/nvidia0 --device /dev/nvidiactl:/dev/nvidiactl --device /dev/nvidia-uvm:/dev/nvidia-uvm"
	}

	templateFile := "../templates/elastic_thought.service.template"

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
