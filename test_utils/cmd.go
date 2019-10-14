package test_utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

type Command struct {
	cmd          string
	args         []string
	dir          string
	timeout      string
	pollInterval string
	env          map[string]string
}

func Cmd(cmd string, args ...string) Command {
	return Command{
		cmd:          cmd,
		args:         args,
		timeout:      "30s",
		pollInterval: "1s",
		env:          map[string]string{},
	}
}

func (c Command) WithDir(dir string) Command {
	return Command{
		cmd:          c.cmd,
		args:         c.args,
		dir:          dir,
		timeout:      c.timeout,
		pollInterval: c.pollInterval,
		env:          c.env,
	}
}

func (c Command) WithTimeout(timeout string) Command {
	return Command{
		cmd:          c.cmd,
		args:         c.args,
		dir:          c.dir,
		timeout:      timeout,
		pollInterval: c.pollInterval,
		env:          c.env,
	}
}

func (c Command) WithEnv(key, val string) Command {
	newEnv := map[string]string{}
	for k, v := range c.env {
		newEnv[k] = v
	}
	newEnv[key] = val
	return Command{
		cmd:          c.cmd,
		args:         c.args,
		dir:          c.dir,
		timeout:      c.timeout,
		pollInterval: c.pollInterval,
		env:          newEnv,
	}
}

func (c Command) Run() *gexec.Session {
	session, err := gexec.Start(c.build(), GinkgoWriter, GinkgoWriter)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	EventuallyWithOffset(1, session, c.timeout, c.pollInterval).Should(gexec.Exit())
	return session
}

func (c Command) build() *exec.Cmd {
	command := exec.Command(c.cmd, c.args...)
	command.Env = os.Environ()
	for k, v := range c.env {
		command.Env = append(command.Env, fmt.Sprintf("%s=%s", k, v))
	}
	if c.dir != "" {
		cwd, err := os.Getwd()
		ExpectWithOffset(2, err).NotTo(HaveOccurred())
		command.Dir = filepath.Join(cwd, c.dir)
	}
	return command
}
