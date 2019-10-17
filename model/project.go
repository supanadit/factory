package model

import (
	"fmt"
	"github.com/naoina/toml"
	"io/ioutil"
	"log"
)

type Project struct {
	ProjectName string
	Alias       string
	Git         Git
	Path        string
}

func (project Project) Save(configuration Configuration) bool {
	var allProject = GetAllProjectConfiguration(configuration)
	allProject.Project = append(allProject.Project, project)
	dataToml, _ := toml.Marshal(&allProject)
	success := true
	exist, _ := project.ExistByAlias(configuration)
	if !exist {
		err := ioutil.WriteFile(configuration.GetProjectConfigFilePath(), dataToml, 0644)
		if err != nil {
			if DEBUG {
				log.Println(err)
			} else {
				fmt.Println("Cannot create configuration file for project")
			}
			success = false
		}
	} else {
		fmt.Println("This Project is Exist")
	}
	return success
}

func (project Project) ExistByAlias(configuration Configuration) (bool, Project) {
	var allProject = GetAllProjectConfiguration(configuration)
	found := false
	var projectFounded Project
	for _, element := range allProject.Project {
		if element.Alias == project.Alias {
			found = true
			projectFounded.Alias = element.Alias
			projectFounded.ProjectName = element.ProjectName
			projectFounded.Path = element.Path
			projectFounded.Git = element.Git
		}
	}
	return found, projectFounded
}

func (project Project) FillFromAlias(configuration Configuration) (bool, Project) {
	exist, projectData := project.ExistByAlias(configuration)
	if exist {
		project.Git = projectData.Git
		project.Path = projectData.Path
		project.ProjectName = projectData.ProjectName
		project.Alias = projectData.Alias
	}
	return exist, project
}

func (project Project) Remove(configuration Configuration) {
	var allProject = GetAllProjectConfiguration(configuration)
	var newProjectConfiguration ProjectConfiguration
	for _, element := range allProject.Project {
		if element != project {
			newProjectConfiguration.Project = append(newProjectConfiguration.Project, element)
		}
	}
	dataToml, _ := toml.Marshal(&newProjectConfiguration)
	err := ioutil.WriteFile(configuration.GetProjectConfigFilePath(), dataToml, 0644)
	if err != nil {
		if DEBUG {
			log.Println(err)
		} else {
			fmt.Println("Failed to remove Project")
		}
	} else {
		fmt.Println("Project Removed")
	}
}
