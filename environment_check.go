package elasticthought

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/couchbaselabs/logg"
	"github.com/tleyden/cbfs/client"
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
	cbfs, err := cbfsclient.New(config.CbfsUrl)
	if err != nil {
		return err
	}

	// write to random cbfs file
	options := cbfsclient.PutOptions{
		ContentType: "text/plain",
	}

	buffer := bytes.NewBuffer([]byte(content))

	if err := cbfs.Put("", destPath, buffer, options); err != nil {
		return fmt.Errorf("Error writing %v to cbfs: %v", destPath, err)
	}

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

	// delete contents on cbfs
	if err := cbfs.Rm(destPath); err != nil {
		return err
	}
	return nil

}

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
