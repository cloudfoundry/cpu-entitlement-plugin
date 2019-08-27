package test_utils

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type Command struct {
	cmd     string
	args    []string
	dir     string
	timeout string
}

func Cmd(cmd string, args ...string) Command {
	return Command{
		cmd:     cmd,
		args:    args,
		timeout: "2s",
	}
}

func (c Command) WithDir(dir string) Command {
	return Command{
		cmd:     c.cmd,
		args:    c.args,
		dir:     dir,
		timeout: c.timeout,
	}
}

func (c Command) WithTimeout(timeout string) Command {
	return Command{
		cmd:     c.cmd,
		args:    c.args,
		dir:     c.dir,
		timeout: timeout,
	}
}

func (c Command) Run() *gexec.Session {
	session, err := gexec.Start(c.build(), GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	EventuallyWithOffset(1, session, c.timeout).Should(gexec.Exit())
	return session
}

func (c Command) build() *exec.Cmd {
	command := exec.Command(c.cmd, c.args...)
	if c.dir != "" {
		cwd, err := os.Getwd()
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		command.Dir = filepath.Join(cwd, c.dir)
	}
	return command
}
