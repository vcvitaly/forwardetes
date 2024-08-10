package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestParsePortFromPropertyLine(t *testing.T) {
	testCases := []struct {
		name         string
		propertyLine string
		expectedPort int
		portsExists  bool
	}{
		{
			name:         "APortIsFound",
			propertyLine: "url1: localhost:8081",
			expectedPort: 8081,
			portsExists:  true,
		},
		{
			name:         "APortIsNotFound",
			propertyLine: "url: jdbc:mysql://localhost:3306/db",
			expectedPort: 0,
			portsExists:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			port, b, err := parsePortFromPropertyLine(tc.propertyLine)
			if err != nil {
				t.Fatal(err)
			}

			if b != tc.portsExists {
				t.Errorf("Expected the match status to be %q, but got %q instead", strconv.FormatBool(tc.portsExists), strconv.FormatBool(b))
			}

			if port != tc.expectedPort {
				t.Errorf("Expected the port to be %d, but got %d instead", tc.expectedPort, port)
			}
		})
	}
}

func TestReadPortsFromAppFile(t *testing.T) {
	testCases := []struct {
		name          string
		appFile       string
		expectedPorts []int
	}{
		{
			name:          "ServicePortsAreParsedFromAppFile",
			appFile:       "testdata/app-test.yml",
			expectedPorts: []int{8081, 8082},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ports, err := parsePortsFromAppFile(tc.appFile)
			if err != nil {
				t.Fatal(err)
			}

			if len(ports) != len(tc.expectedPorts) {
				t.Errorf("Expected %d ports, got %d instead\n", len(tc.expectedPorts), len(ports))
			}

			if !reflect.DeepEqual(ports, tc.expectedPorts) {
				t.Errorf("Expected args %q, got %q istead", intSliceToString(tc.expectedPorts), intSliceToString(ports))
			}
		})
	}
}

func TestParseSvcPortMapping(t *testing.T) {
	testCases := []struct {
		name             string
		mappingsFile     string
		expectedMappings map[int]svcPortMapping
	}{
		{
			name:         "SvcPortsMappingsAreParsed",
			mappingsFile: "testdata/mapping.properties",
			expectedMappings: map[int]svcPortMapping{
				8081: {
					svcName:       "a-service1",
					localPort:     8081,
					containerPort: 8080,
				},
				8082: {
					svcName:       "a-service2",
					localPort:     8082,
					containerPort: 8080,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			allMappings, err := parseSvcPortMapping(tc.mappingsFile)
			if err != nil {
				t.Fatal(err)
			}

			if len(allMappings) != len(tc.expectedMappings) {
				t.Errorf("Expected %d mappings, got %d instead", len(tc.expectedMappings), len(allMappings))
			}

			if !reflect.DeepEqual(allMappings, tc.expectedMappings) {
				t.Errorf("Expected args %q, got %q istead", fmt.Sprint(tc.expectedMappings), fmt.Sprint(allMappings))
			}
		})
	}
}

func intSliceToString(values []int) string {
	valuesText := []string{}

	// Create a string slice using strconv.Itoa.
	// ... Append strings to it.
	for i := range values {
		number := values[i]
		text := strconv.Itoa(number)
		valuesText = append(valuesText, text)
	}

	// Join our string slice.
	result := strings.Join(valuesText, ",")

	return result
}
