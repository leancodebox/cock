package jobmanager

type JobStatus struct {
	Status RunStatus `json:"status"`
	Name   string    `json:"name"`
}

func JobList() []JobStatus {
	var jobNameList []JobStatus
	jobManager.Range(func(key, value any) bool {

		if data, ok := value.(*jobHandle); ok {
			if jobId, ok := key.(string); ok {
				jobNameList = append(jobNameList, JobStatus{Name: jobId, Status: data.status})
			}
		}

		return true
	})
	return jobNameList
}
