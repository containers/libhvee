//go:build windows
// +build windows

package main

import (
	"fmt"
	"time"

	"github.com/containers/libhvee/pkg/wmiext"
)

func main() {
	var service *wmiext.Service
	var err error

	if service, err = wmiext.NewLocalService("root\\virtualization\\v2"); err != nil {
		panic(err)
	}
	defer service.Close()

	item, err := service.SpawnInstance("Msvm_KvpExchangeDataItem")
	if err != nil {
		panic(err)
	}
	defer item.Close()

	_ = item.Put("Name", "jkey-"+fmt.Sprintf("%d", time.Now().Unix()))
	_ = item.Put("Data", "jval-"+fmt.Sprintf("%d", time.Now().Unix()))
	_ = item.Put("Source", 0)

	itemStr := item.GetCimText()
	fmt.Println(itemStr)

	vmms, err := service.GetSingletonInstance("Msvm_VirtualSystemManagementService")
	defer vmms.Close()
	if err != nil {
		panic(err)
	}

	const wql = "Select * From Msvm_ComputerSystem Where ElementName='%s'"

	computerSystem, err := service.FindFirstInstance(fmt.Sprintf(wql, "New Virtual Machine"))
	if err != nil {
		panic(err)
	}
	defer computerSystem.Close()

	var job *wmiext.Instance

	e := vmms.BeginInvoke("AddKvpItems").
		In("TargetSystem", computerSystem).
		In("DataItems", []string{itemStr}).
		Execute()

	fmt.Printf("pre-get %v\n", e)

	fmt.Printf("Pointer -> %d", &job)

	e.Out("Job", &job)

	_ = e.End()
	if err != nil {
		panic(err)
	}

	for {
		status, _, _, err := job.GetAsAny("JobState")
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v\n", status)
		time.Sleep(100 * time.Millisecond)
		job, _ = service.RefetchObject(job)
		if status.(int32) == 7 {
			break
		}
	}
}
