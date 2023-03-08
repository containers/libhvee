//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/containers/libhvee/pkg/hypervctl"
)

func main() {
	var err error

	vmms := hypervctl.VirtualMachineManager{}

	summs, err := vmms.GetSummaryInformation(hypervctl.SummaryRequestNearAll)
	if err != nil {
		fmt.Printf("Could not retrieve virtual machine summaries: %s\n", err.Error())
		os.Exit(1)
	}

	b, err := json.MarshalIndent(summs, "", "\t")

	if err != nil {
		fmt.Println("Failed to generate output")
		os.Exit(1)
	}

	fmt.Println(string(b))
}