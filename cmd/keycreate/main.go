//go:build windows
// +build windows

package main

import (
	"fmt"
	"time"

	"github.com/n1hility/hypervctl/pkg/wmiext"

	"github.com/drtimf/wmi"
)

func main() {
	var service *wmi.Service
	var err error

	if service, err = wmi.NewLocalService("root\\virtualization\\v2"); err != nil {
		panic(err)
	}
	defer service.Close()

	item, err := wmiext.SpawnInstance(service, "Msvm_KvpExchangeDataItem")
	if err != nil {
		panic(err)
	}
	defer item.Close()

	_ = item.Put("Name", "jkey-"+fmt.Sprintf("%d", time.Now().Unix()))
	_ = item.Put("Data", "jval-"+fmt.Sprintf("%d", time.Now().Unix()))
	_ = item.Put("Source", 0)

	itemStr := wmiext.GetCimText(item)
	fmt.Println(itemStr)

	vmms, err := wmiext.GetSingletonInstance(service, "Msvm_VirtualSystemManagementService")
	defer vmms.Close()
	if err != nil {
		panic(err)
	}

	const wql = "Select * From Msvm_ComputerSystem Where ElementName='%s'"

	computerSystem, err := wmiext.FindFirstInstance(service, fmt.Sprintf(wql, "New Virtual Machine"))
	if err != nil {
		panic(err)
	}
	defer computerSystem.Close()

	var job *wmi.Instance

	e := wmiext.BeginInvoke(service, vmms, "AddKvpItems").
		Set("TargetSystem", computerSystem).
		Set("DataItems", []string{itemStr}).
		Execute()

	fmt.Printf("pre-get %v\n", e)

	fmt.Printf("Pointer -> %d", &job)

	e.Get("Job", &job)

	_ = e.End()
	if err != nil {
		panic(err)
	}

	for {
		status, _, _, err := job.Get("JobState")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", status)
		time.Sleep(100 * time.Millisecond)
		job, _ = wmiext.RefetchObject(service, job)
		if status.(int32) == 7 {
			break
		}
	}
}
