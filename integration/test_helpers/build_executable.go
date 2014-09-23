package test_helpers

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var PATExec string

func BuildExecutable() {
	var err error
	PATExec, err = Build("./../../pat")
	Ω(err).ShouldNot(HaveOccurred())
}

func RunPAT(args ...string) *Session {
	session := RunCommand(PATExec, args...)
	return session
}

func RunCommand(cmd string, args ...string) *Session {
	command := exec.Command(cmd, args...)
	session, err := Start(command, GinkgoWriter, GinkgoWriter)
	Ω(err).ShouldNot(HaveOccurred())
	return session
}
