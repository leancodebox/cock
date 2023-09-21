package main

import (
	"fmt"
	"github.com/leancodebox/cock/jobmanager"
	"github.com/leancodebox/cock/jobmanagerserver"
	"log/slog"
	"os"
	"os/signal"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	fileData, err := os.ReadFile("jobConfig.json")
	if err != nil {
		fmt.Println(err)
		return
	}

	// 将所有的任务加入map中
	jobmanagerserver.ServeRun()

	jobmanager.Reg(fileData)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	jobmanagerserver.ServeStop()
	slog.Info("bye~~👋👋")
}
