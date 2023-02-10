//go:build windows
// +build windows

package wmiext

import (
	"fmt"
	"time"

	"github.com/drtimf/wmi"
)

type JobError struct {
	ErrorCode int
}

func (err *JobError) Error() string {
	return fmt.Sprintf("Job failed with error code: %d", err.ErrorCode)
}

func WaitJob(service *wmi.Service, job *wmi.Instance) error {
	for {
		state, _, _, err := job.Get("JobState")
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
		job, _ = RefetchObject(service, job)
		if state.(int32) >= 7 {
			break
		}
	}

	result, _, _, err := job.Get("ErrorCode")
	if err != nil {
		return err
	}

	if result.(int32) != 0 {
		return &JobError{ErrorCode: int(result.(int32))}
	}

	return nil
}
