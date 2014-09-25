package workloads

import (
	"os/user"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/pat/context"
)

type WorkloadAdder interface {
	AddWorkloadStep(WorkloadStep)
}

type WorkloadStep struct {
	Name        string
	Fn          func(context context.Context) error
	Description string
}

type WorkloadList struct {
	Workloads []WorkloadStep
}

var restContext = NewRestWorkload()

func DefaultWorkloadList() *WorkloadList {
	return &WorkloadList{[]WorkloadStep{
		StepWithContext("rest:target", restContext.Target, "Sets the CF target"),
		StepWithContext("rest:login", restContext.Login, "Performs a login to the REST api. This option requires rest:target to be included in the list of workloads"),
		StepWithContext("rest:push", restContext.Push, "Pushes an application using the REST api. This option requires both rest:target and rest:login to be included in the list of workloads"),
		StepWithContext("cf:push", Push, "Pushes an application using the CF command-line"),
		StepWithContext("cf:delete", Delete, "Deletes the most recently pushed app."),
		StepWithContext("cf:generateAndPush", GenerateAndPush, "Generates and pushes a unique application using the CF command-line"),
		StepWithContext("dummy", Dummy, "An empty workload that can be used when a CF environment is not available"),
		StepWithContext("dummyDelete", DummyDelete, "An empty workload that simulates Delete"),
		StepWithContext("dummyWithErrors", DummyWithErrors, "An empty workload that generates errors. This can be used when a CF environment is not available"),
	}}
}

func Step(name string, fn func() error, description string) WorkloadStep {
	return WorkloadStep{name, func(ctx context.Context) error { return fn() }, description}
}

func StepWithContext(name string, fn func(context.Context) error, description string) WorkloadStep {
	return WorkloadStep{name, fn, description}
}

func (self *WorkloadList) DescribeWorkloads(to WorkloadAdder) {
	for _, workload := range self.Workloads {
		to.AddWorkloadStep(workload)
	}
}

func PopulateAppContext(appPath string, manifestPath string, ctx context.Context) error {
	normalizedAppPath, err := normalizePath(appPath)
	if err != nil {
		return err
	}
	ctx.PutString("app", normalizedAppPath)

	normalizedManifestPath, err := normalizePath(manifestPath)
	if err != nil {
		return err
	}
	ctx.PutString("app:manifest", normalizedManifestPath)

	return nil
}

func normalizePath(aPath string) (string, error) {
	if aPath == "" {
		return "", nil
	}

	normalizedPath := filepath.Clean(aPath)
	normalizedPath = filepath.ToSlash(normalizedPath)
	dirs := strings.Split(normalizedPath, "/")

	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	for i, dir := range dirs {
		if dir == "~" {
			dirs[i] = usr.HomeDir
		}
	}

	return filepath.Join(strings.Join(dirs, "/")), nil
}
