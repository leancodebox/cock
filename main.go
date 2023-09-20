package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gen2brain/beeep"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/cock/jobmanager"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	// 将所有的任务加入map中
	serveRun()

	jobmanager.Reg(fileData)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	serveStop()
	slog.Info("bye~~👋👋")
}

var srv *http.Server

func serveRun() *http.Server {
	r := gin.Default()
	srv = &http.Server{
		Addr:           ":9090",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	r.GET("/job-list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": jobmanager.JobList(),
		})
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	return srv
}
func serveStop() {
	if srv == nil {
		return
	}
	slog.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Info("Server Shutdown:", "err", err.Error())
	}
}
