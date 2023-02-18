//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/n1hility/hypervctl/pkg/hypervctl"
)

func main() {
	var err error

	vmms := hypervctl.VirtualMachineManager{}

	vms, err := vmms.GetAll()
	if err != nil {
		fmt.Printf("Could not retrieve virtual machines : %s\n", err.Error())
		os.Exit(1)
	}

	b, err := json.MarshalIndent(vms, "", "\t")

	if err != nil {
		fmt.Println("Failed to generate output")
		os.Exit(1)
	}

	fmt.Println(string(b))
}
