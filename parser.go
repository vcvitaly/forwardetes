package main

import (
	"bufio"
	"fmt"
	"github.com/magiconair/properties"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type svcPortMapping struct {
	svcName       string
	localPort     localPort
	containerPort int
}

type localPort int

func (s svcPortMapping) String() string {
	return fmt.Sprintf("{'%s', '%d', '%d'}", s.svcName, s.localPort, s.containerPort)
}

func parsePortsFromAppFile(appFile string) ([]int, error) {
	file, err := os.Open(appFile)
	if err != nil {
		return nil, fmt.Errorf("an error while opening the app file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("An error while closing the app file: %v", err)
		}
	}(file)

	var ports []int

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		port, portExists, err := parsePortFromPropertyLine(scanner.Text())
		if err != nil {
			return nil, err
		}
		if portExists {
			ports = append(ports, port)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("an error while scanning the app file: %w", err)
	}

	return ports, nil
}

func parsePortFromPropertyLine(propertyLine string) (int, bool, error) {
	re := regexp.MustCompile("localhost:(8\\d{3})")
	match := re.FindStringSubmatch(propertyLine)
	if len(match) < 1 || match[1] == "" {
		return 0, false, nil
	}

	portStr := match[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, false, fmt.Errorf("an error while parsing %q: %w", portStr, err)
	}

	return port, true, nil
}

func parseSvcPortMapping(mappingsFile string) (map[localPort]svcPortMapping, error) {
	p, err := properties.LoadFile(mappingsFile, properties.UTF8)
	if err != nil {
		return nil, err
	}

	allMappingsByLocalPort := make(map[localPort]svcPortMapping)

	for k, v := range p.Map() {
		aLocalPortInt, err := strconv.Atoi(k)
		aLocalPort := localPort(aLocalPortInt)
		if err != nil {
			return nil, fmt.Errorf("could not parse the local port %s: %w", k, err)
		}

		svcNameAndContainerPortSlice := strings.Split(v, ",")
		if len(svcNameAndContainerPortSlice) != 2 {
			return nil, fmt.Errorf("expected 2 elements to be parsed from the string value for local port %d,"+
				"but got %d instead", aLocalPort, len(svcNameAndContainerPortSlice))
		}

		containerPort, err := strconv.Atoi(svcNameAndContainerPortSlice[1])
		if err != nil {
			return nil, fmt.Errorf("could not parse the container port %s: %w", svcNameAndContainerPortSlice[1], err)
		}

		mapping := svcPortMapping{
			svcName:       svcNameAndContainerPortSlice[0],
			localPort:     aLocalPort,
			containerPort: containerPort,
		}
		allMappingsByLocalPort[aLocalPort] = mapping
	}

	return allMappingsByLocalPort, nil
}
