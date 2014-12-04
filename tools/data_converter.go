package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kingpin"
	"github.com/couchbaselabs/logg"
	et "github.com/tleyden/elastic-thought"
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

func doFlat2folder(sourceDir, labelIndex, destDir string) {

	// validations: sourceDir and index exist
	validatePathExists(sourceDir)
	validatePathExists(destDir)
	validatePathExists(labelIndex)

	// Process a single row in labelIndex (examples/mnist/mnist_test_files/1.png 2)
	labelIndexLineHandler := func(line string) error {

		// find the label (eg, 2)
		filename, label, err := findComponentsFromIndexRow(line)
		if err != nil {
			return err
		}

		// create dir for label under sourceDir
		destPath := path.Join(destDir, label)
		if err := et.Mkdir(destPath); err != nil {
			return err
		}

		// strip off any path information from file
		_, filenameNoPath := filepath.Split(filename)

		// source file
		sourceFile := path.Join(sourceDir, filenameNoPath)

		// dest file
		destFile := path.Join(destPath, filenameNoPath)

		// copy file from sourceDir to destDir/label
		et.CopyFileContents(sourceFile, destFile)

		logg.LogTo("DATA_CONVERTER", "Copied %v -> %v", sourceFile, destFile)

		return nil
	}

	// process all rows in labelIndex
	if err := processLabelIndex(labelIndex, labelIndexLineHandler); err != nil {
		log.Fatal(err)
	}

}

// process each line in labelIndex with lineHandler func
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

// example line: "examples/mnist/mnist_test_files/1.png 2"
func findComponentsFromIndexRow(line string) (string, string, error) {

	components := strings.Split(line, " ")
	if len(components) != 2 {
		return "", "", fmt.Errorf("Expected 2 components in %v", line)
	}
	filename := components[0]
	label := components[1]

	return filename, label, nil
}

func validatePathExists(path string) {
	_, err := os.Stat(path)
	if err != nil {
		logg.LogPanic("%v does not exist", path)
	}
}
