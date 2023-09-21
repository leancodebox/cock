package jobmanager

import (
	"encoding/json"
	"fmt"
	"github.com/leancodebox/cock/cockSay"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"time"
)

var startTime = time.Now()

type RunStatus int

const (
	Stop RunStatus = iota
	Running
)

var runStatusName = [...]string{
	"停止",
	"运行",
}

func (d RunStatus) String() string {
	if d < Stop || d > Running {
		return "Unknown"
	}
	return runStatusName[d]
}

const maxExecutionTime = 10 * time.Second // 最大允许的运行时间
const maxConsecutiveFailures = 3          // 连续失败次数的最大值

type OutputType int

const (
	OutputTypeStd  OutputType = iota // 输出到标准输入输出
	OutputTypeFile                   // 输出到文件
)

type RunOptions struct {
	OutputType  OutputType `json:"outputType"`  // 输出方式
	OutputPath  string     `json:"outputPath"`  // 输出路径
	MaxFailures int        `json:"maxFailures"` // 最大失败次数
}

type Job struct {
	JobName string     `json:"jobName"`
	BinPath string     `json:"binPath"`
	Params  []string   `json:"params"`
	Dir     string     `json:"dir"`
	Timer   bool       `json:"timer"`
	Spec    string     `json:"spec"`
	Run     bool       `json:"run"`
	Options RunOptions `json:"options"` // 运行选项
}

type JobConfig struct {
	ResidentTask  []Job `json:"residentTask"`
	ScheduledTask []Job `json:"scheduledTask"`
}

var jobManager sync.Map

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

func Reg(fileData []byte) {

	var jobConfig JobConfig
	err := json.Unmarshal(fileData, &jobConfig)
	if err != nil {
		slog.Info(err.Error())
		return
	}

	for id, job := range jobConfig.ResidentTask {
		jh := jobHandle{jobConfig: job}
		jobManager.Store(fmt.Sprintf("%v%v", id, jh.jobConfig.JobName), &jh)
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
