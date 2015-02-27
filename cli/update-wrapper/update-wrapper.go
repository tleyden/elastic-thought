package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/tleyden/go-etcd/etcd"
)

// This makes it easy to run a command after possibly first grabbing the latest version
// via "go get".  It attempts to solve the problem of having to rebuild docker images
// after tiny code changes, which makes the code-test loop too long.

const (
	KEY_ENABLE_CODE_REFRESH = "/elastic-thought/update-wrapper"
	REPO_URL                = "github.com/tleyden/elastic-thought/..."
)

func main() {

	// connect to etcd and see if we even need to update the code
	requiresUpdate := checkUpdateRequired()

	if requiresUpdate {
		log.Printf("Update-Wrapper: updating to latest code")
		updateAndRebuild(REPO_URL)
	} else {
		log.Printf("Update-Wrapper: skipping update")
	}

	// invoke appropriate binary with command line args
	if err := invokeTarget(); err != nil {
		log.Fatalf("Error invoking target: %v", err)
	}

}

func checkUpdateRequired() bool {

	// find etcd server list from args
	etcdServers := findEtcdServersFromArgs(os.Args)

	// create etcd client
	etcdClient := etcd.NewClient(etcdServers)
	etcdClient.SetConsistency(etcd.STRONG_CONSISTENCY)

	_, err := etcdClient.Get(KEY_ENABLE_CODE_REFRESH, false, false)
	if err != nil {
		// if we got an error, assume key not there
		return false
	}

	return true

}

func findEtcdServersFromArgs(args []string) []string {

	for i, arg := range args {
		if strings.Contains(arg, "--etcd-servers") {
			if strings.Contains(arg, "=") {
				// tokenize on = ..
				tokens := strings.Split(arg, "=")
				commaSepList := tokens[1] // srv1,srv2..
				return strings.Split(commaSepList, ",")
			} else {
				// assume the next arg is list of etcd servers
				if len(args) >= i+1 {
					nextArg := args[i+1]
					// split on "," and return array
					return strings.Split(nextArg, ",")
				}

			}

		}

	}

	return []string{}

}

// update the given repoUrl (github.com/tleyden/elastic-thought/...)
func updateAndRebuild(repoUrl string) {

	// go get -u -v ..
	goGetArgs := []string{
		"get",
		"-u",
		"-v",
		repoUrl,
	}
	cmdGoGet := exec.Command("go", goGetArgs...)
	cmdGoGet.Stdout = os.Stdout
	cmdGoGet.Stderr = os.Stderr

	if err := cmdGoGet.Run(); err != nil {
		log.Printf("Error trying to go get: %v.  Ignoring it", err)
		return
	}

}

func invokeTarget() error {

	// get args with this binary stripped off
	args := os.Args[1:]

	if len(args) == 0 {
		return fmt.Errorf("No target given in args")
	}

	target := args[0]

	remainingArgs := args[1:]

	cmd := exec.Command(target, remainingArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
