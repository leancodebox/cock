package jobmanager

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"
)

type runStatus int

const (
	Stop runStatus = iota
	Running
)

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
	status    runStatus
	confLock  sync.Mutex
	cmd       *exec.Cmd
}

// RunJob 初始化并执行
func (itself *jobHandle) RunJob() {
	itself.confLock.Lock()
	defer itself.confLock.Unlock()
	if itself.status == Running {
		return
	}
	if itself.cmd == nil {
		job := itself.jobConfig
		cmd := exec.Command(job.BinPath, job.Params...)
		cmd.Dir = job.Dir
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		itself.cmd = cmd
	}
	go itself.jobGuard()
}

func (itself *jobHandle) jobGuard() {
	job := itself.jobConfig
	if job.Options.OutputType == OutputTypeFile && job.Options.OutputPath != "" {
		err := os.MkdirAll(job.Options.OutputPath, os.ModePerm)
		if err != nil {
			fmt.Println(err)
		}
		logFile, err := os.OpenFile(job.Options.OutputPath+"/"+job.JobName+"_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println(err)
		}
		defer logFile.Close()
		itself.cmd.Stdout = logFile
		itself.cmd.Stderr = logFile
	}
	counter := 1
	consecutiveFailures := 1
	for {
		startTime := time.Now()
		counter += 1
		cmdErr := itself.cmd.Start()
		if cmdErr == nil {
			itself.status = Running
			cmdErr = itself.cmd.Wait()
			itself.status = Stop
		}
		executionTime := time.Since(startTime)
		if cmdErr != nil {
			fmt.Println(cmdErr)
		}
		if executionTime <= maxExecutionTime {
			consecutiveFailures += 1
		} else {
			consecutiveFailures = 1
		}

		if consecutiveFailures >= max(maxConsecutiveFailures, job.Options.MaxFailures) {
			fmt.Println(job.JobName + "程序连续3次启动失败，停止重启")
			break
		} else {
			fmt.Println(job.JobName + "程序终止尝试重新运行")
		}
	}
}

func (itself *jobHandle) StopJob() {
	itself.confLock.Lock()
	defer itself.confLock.Unlock()
	if itself.cmd == nil && itself.cmd.Process != nil {
		err := itself.cmd.Process.Kill()
		if err != nil {
			slog.Info(err.Error())
			return
		}
	}
}

func reg() {

	fileData, err := os.ReadFile("jobConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var jobConfig JobConfig
	err = json.Unmarshal(fileData, &jobConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	for id, job := range jobConfig.ResidentTask {
		jobManager.Store(fmt.Sprintf("%v%v", id, job.JobName), job)
	}

	// 将所有的任务加入map中

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
