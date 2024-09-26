package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
)

const kubectlCmdName = "kubectl"

func main() {
	if err := checkOS(); err != nil {
		log.Fatalf("An error: %v", err)
	}
	_, err := exec.LookPath(kubectlCmdName)
	if err != nil {
		log.Fatal(err)
	}

	inputFile := flag.String("i", "", "Input file with a list of port-forward commands")
	flag.Parse()

	if *inputFile == "" {
		if err := usage(); err != nil {
			log.Fatalf("An error while printing usage info to stderr: %v", err)
		}
	}

	if err := run(*inputFile, os.Stdout); err != nil {
		log.Fatalf("An error in the run method: %v", err)
	}

	fmt.Println("Quitting")
}

func checkOS() error {
	if runtime.GOOS != "windows" {
		return ErrUnsupportedOs
	}

	return nil
}

func usage() error {
	_, err := fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	if err != nil {
		return err
	}
	flag.PrintDefaults()
	return nil
}

func run(inputFile string, out io.Writer) error {
	allMappings, err := parseSvcPortMapping(inputFile)
	if err != nil {
		return err
	}

	mappingsWithoutOpenPorts := filterOutOpenPortMappings(allMappings)

	params := provideParams(mappingsWithoutOpenPorts, "default")

	err = portForwardAll(params, out)
	if err != nil {
		return err
	}

	return nil
}

func filterOutOpenPortMappings(matchedMappings []svcPortMapping) []svcPortMapping {
	var closedPortMappings []svcPortMapping

	for _, m := range matchedMappings {
		aPortState := scanPort("localhost", int(m.localPort))
		if aPortState.open {
			log.Printf("The port %d on localhost is open by some other process, skipping it", aPortState.port)
			continue
		}

		closedPortMappings = append(closedPortMappings, m)
	}

	return closedPortMappings
}
