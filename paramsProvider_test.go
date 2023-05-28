package main

import (
	"reflect"
	"testing"
)

func TestParamsProvider(t *testing.T) {
	testCases := []struct {
		name         string
		mappings     []svcPortMapping
		namespace    string
		expectedArgs []forwardArgs
	}{
		{
			name: "ArgsAreProvided",
			mappings: []svcPortMapping{
				{
					svcName:       "a-service",
					localPort:     8081,
					containerPort: 8080,
				},
			},
			namespace: "default",
			expectedArgs: []forwardArgs{
				{
					"port-forward",
					"svc/a-service",
					"8081:8080",
					"--namespace",
					"default",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			params := provideParams(tc.mappings, tc.namespace)

			if len(params) != len(tc.expectedArgs) {
				t.Errorf("Expected %d args arrays, got %d instead\n", len(tc.expectedArgs), len(params))
			}

			if len(params[0]) != len(tc.expectedArgs[0]) {
				t.Errorf("Expected %d args, got %d instead", len(tc.expectedArgs[0]), len(params[0]))
			}

			if !reflect.DeepEqual(params[0], tc.expectedArgs[0]) {
				t.Errorf("Expected args %q, got %q istead", tc.expectedArgs[0], params[0])
			}
		})
	}
}
