package main

import (
	"bufio"
	"fmt"
	"log"
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

	sourceDir  = flat2folder.Flag("sourceDir", "Directory with source files").String()
	labelIndex = flat2folder.Flag("labelIndex", "File with file -> label mapping").String()
	destDir    = flat2folder.Flag("destDir", "Destination directory for output folders").String()
)

func init() {
	logg.LogKeys["DATA_CONVERTER"] = true
}

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "flat2folder":
		logg.LogTo("DATA_CONVERTER", "do flat2folder")
		doFlat2folder(*sourceDir, *labelIndex, *destDir)
	default:
		kingpin.UsageErrorf("Invalid / missing command")
	}
}

func processLabelIndex(labelIndex string, lineHandler func(string) error) error {

	// open labelIndex
	file, err := os.Open(labelIndex)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := lineHandler(scanner.Text()); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil

}

func doFlat2folder(sourceDir, labelIndex, destDir string) {

	// validations: sourceDir and index exist
	validatePathExists(sourceDir)
	validatePathExists(labelIndex)

	labelIndexLineHandler := func(line string) error {
		fmt.Println(line)
		return nil
	}

	if err := processLabelIndex(labelIndex, labelIndexLineHandler); err != nil {
		log.Fatal(err)
	}

	// for each row in index (examples/mnist/mnist_test_files/1.png 2)

	// find the label (eg, 2)

	// create dir for label under sourceDir

	// copy file from sourceDir to destDir/label

}

func validatePathExists(path string) {
	_, err := os.Stat(path)
	if err != nil {
		logg.LogPanic("%v does not exist", path)
	}
}
