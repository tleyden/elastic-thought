package elasticthought

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/couchbaselabs/logg"
)

func EnvironmentSanityCheck(config Configuration) error {

	if err := CbfsSanityCheck(config); err != nil {
		return err
	}
	logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity check passed")

	return nil

}

func CbfsReadWriteFile(config Configuration, destPath, content string) error {

	// get cbfs client
	// Create a cbfs client
	cbfs, err := NewBlobStore(config.CbfsUrl)
	if err != nil {
		return err
	}

	// write to random cbfs file
	options := BlobPutOptions{}
	options.ContentType = "text/plain"

	buffer := bytes.NewBuffer([]byte(content))

	if err := cbfs.Put("", destPath, buffer, options); err != nil {
		return fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
	}

	// delete it after we are done
	defer func() {
		if err := cbfs.Rm(destPath); err != nil {
			logg.LogError(fmt.Errorf("Error deleting %v from cbfs", destPath))
		}
	}()

	// read contents from cbfs file
	reader, err := cbfs.Get(destPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	if string(bytes) != content {
		return fmt.Errorf("Content did not match expected")
	}

	// now make sure the file shows up on at least 3 cbfs nodes
	for {
		// read contents from cbfs file
		log.Printf("Getting filehandle of %v", destPath)
		fileHandle, err := cbfs.OpenFile(destPath)
		if err != nil {
			logg.LogError(fmt.Errorf("Error calling OpenFile on %v: %v", destPath, err))
			return err
		}

		nodes := fileHandle.Nodes()
		if len(nodes) >= config.NumCbfsClusterNodes {
			log.Printf("%v present on %v nodes, which is sufficient", destPath, len(nodes))
			return nil
		}

		log.Printf("%v only present on %v nodes, which is < %v", destPath, len(nodes), config.NumCbfsClusterNodes)

		time.Sleep(1 * time.Second)

	}

}

// Perform a
func CbfsSanityCheck(config Configuration) error {

	uuid := NewUuid() // use uuid so other nodes on cluster don't conflict
	numAttempts := 20
	for i := 0; i < numAttempts; i++ {
		filename := fmt.Sprintf("env_check_%v_%v", uuid, i)
		content := fmt.Sprintf("Hello %v_%v", uuid, i)
		err := CbfsReadWriteFile(config, filename, content)
		if err == nil {
			logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity ok: %v", filename)
			return nil
		}
		logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity failed # %v: %v", i, filename)
		if i >= (numAttempts - 1) {
			logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity check giving up")
			return err
		} else {
			logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity check sleeping ..")
			time.Sleep(time.Duration(i) * time.Second)
			logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity check done sleeping")
		}
	}
	return fmt.Errorf("Exhausted attempts")

}
