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

type Job struct {
	JobName string   `json:"jobName"`
	BinPath string   `json:"binPath"`
	Params  []string `json:"params"`
	Dir     string   `json:"dir"`
	Timer   bool     `json:"timer"`
	Spec    string   `json:"spec"`
	Run     bool     `json:"run"`
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
			defer wg.Done()
			if job.Run == false {
				return
			}
			consecutiveFailures := 1
			for {
				startTime := time.Now()
				cmd := exec.Command(job.BinPath, job.Params...)
				cmd.Dir = job.Dir
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
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
				if consecutiveFailures >= maxConsecutiveFailures {
					cockSay(job.JobName + "程序连续3次启动失败，停止重启")
					break
				} else {
					cockSay(job.JobName + "程序终止尝试重新运行")
				}
			}
		}(job)
	}
	cockSay(fmt.Sprintf("cock启动成功"))

	wg.Wait()
	fmt.Println("当前没有任务")
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
				cmdErr := cmd.Run()
				if cmdErr != nil {
					fmt.Println(cmdErr)
				}
				fmt.Println("end")
			}
		}(job))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(job.JobName + "加入定时任务")
	}
	c.Run()
}
