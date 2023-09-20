package main

import (
	"encoding/json"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/robfig/cron/v3"
	"os"
	"os/exec"
	"sync"
	"time"
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

func cockSay(msg string) {
	err := beeep.Notify("Cock", msg, "")
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

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

	go schedule(jobConfig.ScheduledTask)

	wg := sync.WaitGroup{}
	counter := 0
	for _, job := range jobConfig.ResidentTask {
		wg.Add(1)
		go func(job Job) {
			defer func() {
				if p := recover(); p != nil {
					data, _ := json.Marshal(p)
					cockSay("任务崩溃" + string(data))
				}
			}()
			defer wg.Done()
			if job.Run == false {
				return
			}
			fmt.Println(job.JobName + "加入常驻任务")
			consecutiveFailures := 1
			cmd := exec.Command(job.BinPath, job.Params...)
			cmd.Dir = job.Dir
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
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
				cmd.Stdout = logFile
				cmd.Stderr = logFile
			}
			for {
				startTime := time.Now()
				counter += 1
				cmdErr := cmd.Run()
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
					cockSay(job.JobName + "程序连续3次启动失败，停止重启")
					break
				} else {
					fmt.Println(job.JobName + "程序终止尝试重新运行")
					cockSay(job.JobName + "程序终止尝试重新运行")
				}
			}
		}(job)
	}
	cockSay(fmt.Sprintf("cock启动成功"))

	wg.Wait()
	fmt.Println("当前没有任务,常驻任务，退出所有程序")
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
						fmt.Println(err)
					}
					logFile, err := os.OpenFile(job.Options.OutputPath+"/"+job.JobName+"_log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
					if err != nil {
						fmt.Println(err)
					}
					defer logFile.Close()
					cmd.Stdout = logFile
					cmd.Stderr = logFile
				}
				cmdErr := cmd.Run()
				if cmdErr != nil {
					fmt.Println(cmdErr)
				}
			}
		}(job))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(job.JobName + "加入定时任务")
	}
	c.Run()
}
