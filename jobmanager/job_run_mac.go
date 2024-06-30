//go:build darwin || (openbsd && !mips64)

package jobmanager

import (
	"os"
	"os/exec"
)

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
