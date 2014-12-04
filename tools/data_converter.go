package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/couchbaselabs/logg"
)

/*
Tool for converting input data from one form to another
*/

var (
	app         = kingpin.New("data converter", "A command-line tool for converting data.")
	flat2folder = app.Command("flat2folder", "Take a flat list of files + an index and convert to folder labels")

	sourceDir = flat2folder.Flag("sourceDir", "Directory with source files").String()
	index     = flat2folder.Flag("labelIndex", "File with file -> label mapping").String()
	destDir   = flat2folder.Flag("destDir", "Destination directory for output folders").String()
)

func init() {
	logg.LogKeys["DATA_CONVERTER"] = true
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "flat2folder":
		logg.LogTo("DATA_CONVERTER", "do flat2folder")
	default:
		kingpin.UsageErrorf("Invalid / missing command")
	}
}
