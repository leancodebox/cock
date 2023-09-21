package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/cock/jobmanager"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"
)

const maxExecutionTime = 10 * time.Second // 最大允许的运行时间
const maxConsecutiveFailures = 3          // 连续失败次数的最大值

type OutputType int

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	fileData, err := os.ReadFile("jobConfig.json")
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

//go:embed  all:actor/dist/**
var actorFS embed.FS

func serveRun() *http.Server {
	//r := gin.Default()

	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	srv = &http.Server{
		Addr:           ":9090",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	r.Use(GinCors)
	act := r.Group("actor")
	act.StaticFS("", PFilSystem("./actor/dist", actorFS))
	api := r.Group("api")
	api.GET("/job-list", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": jobmanager.JobList(),
		})
	})
	type JobUpdateReq struct {
		JobId string `json:"jobId"`
	}
	api.POST("/run-job", func(c *gin.Context) {
		var params JobUpdateReq
		_ = c.ShouldBind(&params)
		err := jobmanager.JobRun(params.JobId)
		msg := "success"
		if err != nil {
			msg = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{
			"message": msg,
		})
	})
	api.POST("/stop-job", func(c *gin.Context) {
		var params JobUpdateReq
		_ = c.ShouldBind(&params)
		err := jobmanager.JobStop(params.JobId)
		msg := "success"
		if err != nil {
			msg = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{
			"message": msg,
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

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}

func upFsHandle(pPath string, fSys fs.FS) fsFunc {
	return func(name string) (fs.File, error) {
		assetPath := path.Join(pPath, name)
		// If we can't find the asset, fs can handle the error
		file, err := fSys.Open(assetPath)
		if err != nil {
			slog.Error(err.Error())
			return nil, err
		}
		return file, err
	}
}

func PFilSystem(pPath string, fSys fs.FS) http.FileSystem {
	return http.FS(upFsHandle(pPath, fSys))
}

func GinCors(context *gin.Context) {
	method := context.Request.Method
	context.Header("Access-Control-Allow-Origin", "*")
	context.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token, Authorization, Token, New-Token")
	context.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PATCH, PUT")
	context.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type,New-Token")
	context.Header("Access-Control-Allow-Credentials", "true")
	if method == "OPTIONS" {
		context.AbortWithStatus(http.StatusNoContent)
	}
	context.Next()
}
