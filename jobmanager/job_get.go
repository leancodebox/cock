package jobmanager

import (
	"errors"
	"time"
)

type JobStatus struct {
	Status  RunStatus `json:"status"`
	Name    string    `json:"name"`
	OpenRun bool      `json:"openRun"`
	Type    int       `json:"type"` // 1 是job 2是 task
}

func JobList() []JobStatus {
	var jobNameList []JobStatus
	for _, jobId := range jobIdList.getAll() {
		data, ok := jobManager.Load(jobId)
		if !ok {
			continue
		}
		if jh, ok := data.(*jobHandle); ok {
			jobNameList = append(jobNameList, JobStatus{Name: jobId, Status: jh.status, OpenRun: jh.jobConfig.Run, Type: 1})
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

func TaskList() []JobStatus {
	var jobNameList []JobStatus
	for _, jobId := range taskIdList.getAll() {
		data, ok := taskManager.Load(jobId)
		if !ok {
			continue
		}
		if th, ok := data.(*taskHandle); ok {
			jobNameList = append(jobNameList, JobStatus{Name: jobId, Status: th.status, OpenRun: th.jobConfig.Run, Type: 2})
		}
	}
	return jobNameList
}

func getTaskByTaskId(taskId string) *taskHandle {
	if data, ok := taskManager.Load(taskId); ok {
		if th, ok := data.(*taskHandle); ok {
			return th
		}
	}
	return nil
}

func RunTask(taskId string) error {
	task := getTaskByTaskId(taskId)
	if task == nil {
		return errors.New("taskId不存在")
	}
	return task.RunOnce()
}
