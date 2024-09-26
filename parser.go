package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type svcPortMapping struct {
	svcName       string
	localPort     port
	containerPort port
}

type port int

func (s svcPortMapping) String() string {
	return fmt.Sprintf("{'%s', '%d', '%d'}", s.svcName, s.localPort, s.containerPort)
}

func parseSvcPortMapping(mappingsFile string) ([]svcPortMapping, error) {
	lines, err := readLines(mappingsFile)
	if err != nil {
		return nil, err
	}

	allMappings := make([]svcPortMapping, len(lines))

	for _, line := range lines {
		svcAndPortsSlice := strings.Split(line, " ")
		if len(svcAndPortsSlice) != 2 {
			return nil, fmt.Errorf(
				"expected format is svc/<name> <port>:<containerPort>,"+
					"but got %q instead", line,
			)
		}

		ports := svcAndPortsSlice[1]
		portsSlice := strings.Split(ports, ":")

		if len(portsSlice) != 2 {
			return nil, fmt.Errorf(
				"expected format of the ports slice is <port>:<containerPort>,"+
					"but got %q instead", ports,
			)
		}

		aLocalPortInt, err := strconv.Atoi(portsSlice[0])
		if err != nil {
			return nil, fmt.Errorf("could not parse the local port %s: %w", portsSlice[0], err)
		}
		aContainerPortInt, err := strconv.Atoi(portsSlice[1])
		if err != nil {
			return nil, fmt.Errorf("could not parse the container port %s: %w", portsSlice[1], err)
		}

		aLocalPort := port(aLocalPortInt)
		aContainerPort := port(aContainerPortInt)

		mapping := svcPortMapping{
			svcName:       svcAndPortsSlice[0],
			localPort:     aLocalPort,
			containerPort: aContainerPort,
		}
		allMappings = append(allMappings, mapping)
	}

	return allMappings, nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			_ = fmt.Errorf("could not close file %s: %w", path, err)
			os.Exit(1)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") && len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}
