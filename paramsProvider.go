package main

import "fmt"

type forwardArgs []string

func provideParams(svcPortMappings []svcPortMapping, namespace string) []forwardArgs {
	fwArgs := make([]forwardArgs, 0, len(svcPortMappings))

	for _, m := range svcPortMappings {
		var args []string
		args = append(
			args,
			"port-forward", "svc/"+m.svcName,
			fmt.Sprintf("%d", m.localPort)+":"+fmt.Sprintf("%d", m.containerPort),
			"--namespace",
			namespace,
		)
		fwArgs = append(fwArgs, args)
	}

	return fwArgs
}
