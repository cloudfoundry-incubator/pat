package main

import (
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
)

type Workload struct {
	Users int	//how many users should do each workload
	Commands
}

type Commands struct {
	Iterations int	//how many times we should run through each list of commands
	Command []Gcf	//list of actual commands
}

type Gcf struct {
	Iterations int	//how many times we should do each command
	Input struct {
		Cmd	string
		Name	string
	}
}

func ParseWorkload(fName string) (*Workload, error) {
	var workload Workload

	data, err := loadFile(fName)
	if err != nil {
		return nil, err
	}

	workload_data := data.(map[interface{}]interface{})
	commands_data := workload_data["workload"].(map[interface{}]interface{})
	gcf_data := commands_data["commands"].([]interface{})

	workload.Users = workload_data["users"].(int)
	workload.Commands.Iterations = commands_data["loop"].(int)
	for _, gcf_cmd := range gcf_data {
		var cur_gcf Gcf
		gcf_data := gcf_cmd.(map[interface{}]interface{})
		single_cmd := gcf_data["input"].(map[interface{}]interface{})

		cur_gcf.Iterations = gcf_data["loop"].(int)
		cur_gcf.Input.Cmd = single_cmd["cmd"].(string)
		if single_cmd["name"] != nil {
			cur_gcf.Input.Name = single_cmd["name"].(string)
		}
		workload.Commands.Command = append(workload.Commands.Command, cur_gcf)
	}
	return &workload, nil
}

func loadFile(fName string) (interface{}, error) {
	var fLoaded interface{}

	file, err := ioutil.ReadFile(fName)
	if err != nil {
		return nil, err
	}

	err = goyaml.Unmarshal(file, &fLoaded)
	if err != nil {
		return nil, err
	}

	return fLoaded, nil
}

func main() {
	work, err := ParseWorkload("workload.yml")
	if err != nil {
		panic(err)
	}

	fmt.Println(work)
}
