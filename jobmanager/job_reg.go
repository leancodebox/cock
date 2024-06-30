package jobmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/leancodebox/cock/cockSay"
	"github.com/leancodebox/cock/resource"
	"github.com/robfig/cron/v3"
	"log/slog"
	"os"
	"os/exec"
	"path"
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

		if !itself.jobConfig.Run {
			msg := "溜了溜了"
			slog.Info(msg)
			cockSay.Send(msg)
			break
		}

		if consecutiveFailures >= max(maxConsecutiveFailures, job.Options.MaxFailures) {
			msg := job.JobName + "程序连续3次启动失败，停止重启"
			slog.Info(msg)
			cockSay.Send(msg)
			break
		} else {
			msg := job.JobName + "程序终止尝试重新运行"
			cockSay.Send(msg)
			slog.Info(msg)
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

func RegByUserConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("获取家目录失败:", err)
		return err
	}
	fmt.Println("当前系统的家目录:", homeDir)
	configDir := path.Join(homeDir, ".cockTaskConfig")
	if _, err = os.Stat(configDir); os.IsNotExist(err) {
		err = os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	jobConfigPath := path.Join(configDir, "jobConfig.json")
	if _, err = os.Stat(jobConfigPath); os.IsNotExist(err) {
		err = os.WriteFile(jobConfigPath, resource.GetJobConfigDefault(), 0644)
		if err != nil {
			fmt.Println("无法写入文件，错误：", err)
			return err
		}
	}
	fileData, err := os.ReadFile(jobConfigPath)
	if err != nil {
		return err
	}
	Reg(fileData)
	return nil
}

func Reg(fileData []byte) {

	err := json.Unmarshal(fileData, &jobConfig)
	if err != nil {
		slog.Info(err.Error())
		return
	}

	for id, job := range jobConfig.ResidentTask {
		jh := jobHandle{jobConfig: job}
		jobId := fmt.Sprintf("job-%v%v", id, jh.jobConfig.JobName)
		jobManager.Store(jobId, &jh)
		jobIdList.append(jobId)
		jh.RunJob()
		slog.Info(fmt.Sprintf("%v 加入常驻任务", jobId))
	}

	go schedule(jobConfig.ScheduledTask)
}

var c = cron.New()

var taskManager sync.Map
var taskIdList JobListManager

type taskHandle struct {
	jobConfig   Job
	status      RunStatus
	entityId    cron.EntryID
	runOnceLock sync.Mutex
}

func (itself *taskHandle) RunOnce() error {
	if itself.runOnceLock.TryLock() {
		go func(job Job) {
			defer itself.runOnceLock.Unlock()
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
		}(itself.jobConfig)
		return nil
	} else {
		return errors.New("上次手动运行尚未结束")
	}
	return nil
}

func schedule(jobList []Job) {

	for id, job := range jobList {

		if job.Spec == "" {
			slog.Warn("Spec is blank")
			//continue
		}
		// 初始化并记录所有的配置
		th := taskHandle{jobConfig: job}
		taskId := fmt.Sprintf("%v-%v-%v", `task`, id, job.JobName)
		taskIdList.append(taskId)
		taskManager.Store(taskId, &th)
	}

	for _, taskId := range taskIdList.getAll() {
		data, ok := taskManager.Load(taskId)
		if ok == false {
			continue
		}
		th, ok := data.(*taskHandle)
		if ok == false {
			continue
		}
		if th.jobConfig.Run == false {
			continue
		}
		entityId, err := c.AddFunc(th.jobConfig.Spec, func(job Job) func() {
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
		}(th.jobConfig))
		if err != nil {
			slog.Error(err.Error())
		} else {
			th.entityId = entityId
			th.status = Running
			slog.Info(fmt.Sprintf("%v 加入任务", taskId))
		}
	}

	c.Run()
}
