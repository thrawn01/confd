package initd

import (
	"syscall"

	"os/exec"

	"github.com/golang/go/src/pkg/os"
	"github.com/kelseyhightower/confd/log"
	"github.com/kelseyhightower/confd/resource/template"
)

var (
	childPid int
	setsid   bool
)

func Execute(config template.Config) error {
	cmd := exec.Command("/bin/true")
	// Setsid if true
	cmd.SysProcAttr.Setsid = setsid
	// runs fork() and exec()
	if err := cmd.Start(); err != nil {
		return err
	}
	// Capture our child pid
	childPid = cmd.Process.Pid
	return nil
}

func HandleSignal(signum os.Signal) {
	log.Debug("Received signal %d.\n", signum)

	if signum == syscall.SIGTSTP || signum == syscall.SIGTTIN || signum == syscall.SIGTTOU {
		if setsid {
			log.Debug("Running in setsid mode, so forwarding SIGSTOP instead.\n")
			ForwardSignal(syscall.SIGSTOP)
		} else {
			log.Debug("Not running in setsid mode, so forwarding the original signal (%d).\n", signum)
			ForwardSignal(signum)
		}

		log.Debug("Suspending self due to TTY signal.\n")
		os.Kill(syscall.Getpid(), syscall.SIGSTOP)
	} else {
		ForwardSignal(signum)
	}
}

func ForwardSignal(signum int) {
	if childPid <= 0 {
		log.Debug("Didn't forward signal %d, no children exist yet.\n", signum)
		return
	}

	if setsid {
		os.Kill(-childPid)
	} else {
		os.Kill(childPid)
	}
	log.Debug("Forwarded signal %d to children.\n", signum)
}
