package jobmanager

import (
	"encoding/json"
	"fmt"
	"github.com/leancodebox/cock/cockSay"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"sync"
	"time"
)

var startTime = time.Now()

type RunStatus int

const (
	Stop RunStatus = iota
	Running
)


var jobManager sync.Map
var jobIdList JobListManager

type JobListManager struct {
	jobIdList []string
	lock      sync.Mutex
}

func (itself *JobListManager) append(jobId ...string) {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	itself.jobIdList = append(itself.jobIdList, jobId...)
}

func (itself *JobListManager) getAll() []string {
	itself.lock.Lock()
	defer itself.lock.Unlock()
	return slices.Clone(itself.jobIdList)
}

type jobHandle struct {
	jobConfig Job
	status    RunStatus
	confLock  sync.Mutex
	cmd       *exec.Cmd
}

// RunJob 初始化并执行
func (itself *jobHandle) RunJob() {
	itself.confLock.Lock()
	defer itself.confLock.Unlock()
	if itself.cmd == nil {
		job := itself.jobConfig
		cmd := exec.Command(job.BinPath, job.Params...)
		cmd.Dir = job.Dir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		itself.cmd = cmd
	} else {
		return
	}
	go itself.jobGuard()
}

func (itself *jobHandle) ForceRunJob() {
	itself.jobConfig.Run = true
	itself.RunJob()
}

func (itself *jobHandle) jobGuard() {
	defer func() {
		itself.cmd = nil
	}()
	job := itself.jobConfig
	if job.Options.OutputType == OutputTypeFile && job.Options.OutputPath != "" {
		err := os.MkdirAll(job.Options.OutputPath, os.ModePerm)
		if err != nil {
			slog.Info(err.Error())
		}
		logFile, err := os.OpenFile(job.Options.OutputPath+"/"+job.JobName+"_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Info(err.Error())
		}
		defer logFile.Close()
		itself.cmd.Stdout = logFile
		itself.cmd.Stderr = logFile
	}
	counter := 1
	consecutiveFailures := 1
	for {
		if !itself.jobConfig.Run {
			slog.Info("no run ")
			return
		}
		unitStartTime := time.Now()
		counter += 1
		cmdErr := itself.cmd.Start()
		if cmdErr == nil {
			itself.status = Running
			cmdErr = itself.cmd.Wait()
			itself.status = Stop
		}

		executionTime := time.Since(unitStartTime)
		if cmdErr != nil {
			slog.Info(cmdErr.Error())
		}
		if executionTime <= maxExecutionTime {
			consecutiveFailures += 1
		} else {
			consecutiveFailures = 1
		}

		if consecutiveFailures >= max(maxConsecutiveFailures, job.Options.MaxFailures) {
			slog.Info(job.JobName + "程序连续3次启动失败，停止重启")
			cockSay.Send(job.JobName + "程序连续3次启动失败，停止重启")
			break
		} else {
			cockSay.Send(job.JobName + "程序终止尝试重新运行")
			slog.Info(job.JobName + "程序终止尝试重新运行")
		}
	}
}

func (itself *jobHandle) StopJob() {
	itself.confLock.Lock()
	defer itself.confLock.Unlock()
	itself.jobConfig.Run = false
	if itself.cmd != nil && itself.cmd.Process != nil {
		err := itself.cmd.Process.Kill()
		if err != nil {
			slog.Info(err.Error())
			return
		}
		itself.cmd = nil
	}
}

var jobConfig JobConfig

func Reg(fileData []byte) {

	err := json.Unmarshal(fileData, &jobConfig)
	if err != nil {
		slog.Info(err.Error())
		return
	}

	for id, job := range jobConfig.ResidentTask {
		jh := jobHandle{jobConfig: job}
		jobId := fmt.Sprintf("%v%v", id, jh.jobConfig.JobName)
		jobManager.Store(jobId, &jh)
		jobIdList.append(jobId)
		jh.RunJob()
		slog.Info(fmt.Sprintf("%v%v加入常驻任务", id, jh.jobConfig.JobName))
	}

	go schedule(jobConfig.ScheduledTask)
}

func schedule(jobList []Job) {
	var c = cron.New()
	for _, job := range jobList {
		if job.Run == false {
			continue
		}
		if job.Spec == "" {
			continue
		}
		_, err := c.AddFunc(job.Spec, func(job Job) func() {
			return func() {
				cmd := exec.Command(job.BinPath, job.Params...)
				cmd.Dir = job.Dir
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if job.Options.OutputType == OutputTypeFile && job.Options.OutputPath != "" {
					err := os.MkdirAll(job.Options.OutputPath, os.ModePerm)
					if err != nil {
						slog.Info(err.Error())
					}
					logFile, err := os.OpenFile(job.Options.OutputPath+"/"+job.JobName+"_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
					if err != nil {
						slog.Info(err.Error())
					}
					defer logFile.Close()
					cmd.Stdout = logFile
					cmd.Stderr = logFile
				}
				cmdErr := cmd.Run()
				if cmdErr != nil {
					slog.Info(cmdErr.Error())
				}
			}
		}(job))
		if err != nil {
			slog.Info(err.Error())
		}
		slog.Info(job.JobName + "加入定时任务")
	}
	c.Run()
}
