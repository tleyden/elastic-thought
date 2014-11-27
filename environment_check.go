package elasticthought

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/couchbaselabs/cbfs/client"
	"github.com/couchbaselabs/logg"
)

func EnvironmentSanityCheck(config Configuration) error {

	if err := CbfsSanityCheck(config); err != nil {
		return err
	}
	logg.LogTo("ELASTIC_THOUGHT", "Cbfs sanity check passed")

	return nil

}

func CbfsSanityCheck(config Configuration) error {

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

	content := "Hello"
	destPath := "env_check.txt"
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
