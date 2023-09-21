package jobmanager

import (
	"time"
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
type BaseConfig struct {
	Dashboard struct {
		Port int `json:"port"`
	} `json:"dashboard"`
}
type JobConfig struct {
	ResidentTask  []Job      `json:"residentTask"`
	ScheduledTask []Job      `json:"scheduledTask"`
	Config        BaseConfig `json:"config"`
}
