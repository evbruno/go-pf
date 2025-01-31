package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {
	fmt.Printf("=== Running %s ===\n", os.Args[0])

	if len(os.Args) < 3 {
		fmt.Println("Usage: go-pf <k8s-context> <k8s-namespace>")
		os.Exit(1)
	}

	ctx := os.Args[1]
	ns := os.Args[2]

	svcs, err := getKubectlServices(ctx, ns)
	if err != nil {
		panic(err)
	}

	duplicatedPorts := findDuplicatedPorts(svcs)

	for _, svc := range svcs {
		fmt.Println(svc.Name, strings.Join(svc.Ports, "-"))
	}

	slices.Sort(duplicatedPorts)
	fmt.Println("duplicatedPorts:", duplicatedPorts)
}

func findDuplicatedPorts(svcs []K8SService) []string {
	portsInUse := make(map[string]bool)
	duplicatedPorts := make(map[string]bool)

	for _, svc := range svcs {
		for _, port := range svc.Ports {
			if portsInUse[port] {
				duplicatedPorts[port] = true
			} else {
				portsInUse[port] = true
			}
		}
	}

	var result []string
	for port := range duplicatedPorts {
		result = append(result, port)
	}
	return result
}
