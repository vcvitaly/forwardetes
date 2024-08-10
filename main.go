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

	mappingsFile := flag.String("m", "", "Properties file with port=svc key-value pairs")
	appFile := flag.String("a", "", "Application yml or properties file")
	namespace := flag.String("n", "default", "Kubernetes namespace")
	flag.Parse()

	if *mappingsFile == "" || *appFile == "" {
		if err := usage(); err != nil {
			log.Fatalf("An error while printing usage info to stderr: %v", err)
		}
	}

	if err := run(*mappingsFile, *appFile, *namespace, os.Stdout); err != nil {
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

func run(mappingsFile string, appFile string, namespace string, out io.Writer) error {
	allMappingsByLocalPort, err := parseSvcPortMapping(mappingsFile)
	if err != nil {
		return err
	}

	appPorts, err := parsePortsFromAppFile(appFile)
	if err != nil {
		return err
	}

	matchedMappings := findMatchingSvcMappings(allMappingsByLocalPort, appPorts)

	matchedMappingsWithClosedPorts := filterClosedPorts(matchedMappings)

	params := provideParams(matchedMappingsWithClosedPorts, namespace)

	err = portForwardAll(params, out)
	if err != nil {
		return err
	}

	return nil
}

func findMatchingSvcMappings(allMappingsByLocalPort map[localPort]svcPortMapping, appPorts []int) []svcPortMapping {
	var matchedMappings []svcPortMapping

	for _, port := range appPorts {
		if mapping, ok := allMappingsByLocalPort[localPort(port)]; ok {
			matchedMappings = append(matchedMappings, mapping)
		}
	}

	return matchedMappings
}

func filterClosedPorts(matchedMappings []svcPortMapping) []svcPortMapping {
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
