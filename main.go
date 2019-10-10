package main

import (
	"bufio"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/naoina/toml"
	"github.com/ozgio/strutil"
	"github.com/supanadit/devops-factory/model"
	"io/ioutil"
	"os"
	"strings"
)

type args struct {
	New string `arg:"-n,separate" help:"New Project"`
}

func (args) Version() string {
	return "DevOps Factory 0.0.1 Beta"
}

func main() {
	var args args
	fmt.Println("Cross Platform Swiss Army Knife for DevOps")
	arg.MustParse(&args)

	if args.New != "" {
		cfg := model.LoadDefaultConfiguration()
		var project model.Project
		alias := strutil.Slugify(args.New)
		project.ProjectName = args.New
		project.Alias = alias
		project.Path = cfg.GetProjectPath() + "/" + project.Alias
		newProjectPath := project.Path
		if _, err := os.Stat(newProjectPath); os.IsNotExist(err) {
			_ = os.Mkdir(newProjectPath, os.ModePerm)
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("URL Git Repository : ")
		urlGit, _ := reader.ReadString('\n')
		fmt.Print(urlGit)
		var git model.Git
		git.Url = strings.TrimSuffix(urlGit, "\n")
		git.Path = project.Path
		project.Git = git
		var allProject model.ProjectConfiguration
		allProject.Project = append(allProject.Project, project)
		dataToml, _ := toml.Marshal(&allProject)
		//fmt.Printf("%s", dataToml)
		err := ioutil.WriteFile(cfg.GetProjectConfigFilePath(), dataToml, 0644)
		if err != nil {
			panic(err)
		}
	}
}
