package jobmanager

import (
	"errors"
	"time"
)

type JobStatus struct {
	Status  RunStatus `json:"status"`
	Name    string    `json:"name"`
	OpenRun bool      `json:"openRun"`
}

func JobList() []JobStatus {
	var jobNameList []JobStatus
	for _, jobId := range jobIdList.getAll() {
		data, ok := jobManager.Load(jobId)
		if !ok {
			continue
		}
		if jh, ok := data.(*jobHandle); ok {
			jobNameList = append(jobNameList, JobStatus{Name: jobId, Status: jh.status, OpenRun: jh.jobConfig.Run})
		}
	}

	return jobNameList
}

func getJobByJobId(jobId string) *jobHandle {
	if data, ok := jobManager.Load(jobId); ok {
		if jh, ok := data.(*jobHandle); ok {
			return jh
		}
	}
	return nil
}

func JobRun(jobId string) error {
	jh := getJobByJobId(jobId)
	if jh == nil {
		return errors.New("jobId不存在")
	}
	jh.ForceRunJob()
	return nil
}

func JobStop(jobId string) error {
	jh := getJobByJobId(jobId)
	if jh == nil {
		return errors.New("jobId不存在")
	}
	jh.StopJob()
	return nil
}

func RunStartTime() time.Time {
	return startTime
}

func GetHttpConfig() BaseConfig {
	return jobConfig.Config
}
