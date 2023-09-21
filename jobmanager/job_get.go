package jobmanager

import "errors"

type JobStatus struct {
	Status  RunStatus `json:"status"`
	Name    string    `json:"name"`
	OpenRun bool      `json:"openRun"`
}

func JobList() []JobStatus {
	var jobNameList []JobStatus
	jobManager.Range(func(key, value any) bool {

		if data, ok := value.(*jobHandle); ok {
			if jobId, ok := key.(string); ok {
				jobNameList = append(jobNameList, JobStatus{Name: jobId, Status: data.status, OpenRun: data.jobConfig.Run})
			}
		}
		return true
	})
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
