package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

type Job struct {
	JobName string   `json:"jobName"`
	BinPath string   `json:"binPath"`
	Params  []string `json:"params"`
	Dir     string   `json:"dir"`
	Run     bool     `json:"run"`
}

func main() {
	fileData, err := os.ReadFile("jobConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}
	var jobList []Job
	err = json.Unmarshal(fileData, &jobList)
	if err != nil {
		fmt.Println(err)
		return
	}

	wg := sync.WaitGroup{}
	for _, job := range jobList {
		wg.Add(1)
		go func(job Job) {
			defer wg.Done()
			if job.Run == false {
				return
			}

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
		}(job)
	}
	wg.Wait()
}
