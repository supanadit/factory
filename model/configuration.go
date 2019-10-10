package model

import (
	"log"
	"os"
	"os/user"
)

// Constant Value
const DEBUG bool = false
const DirectoryName string = "DevOpsFactory"
const ProjectDirectoryName string = "Project"
const ProjectFileName string = "projects.toml" // Format Must Be TOML

// Default Configuration
type Configuration struct {
	Home      string
	Directory string
	Project   string
}

type ProjectConfiguration struct {
	Project []Project
}

func LoadDefaultConfiguration() Configuration {
	usr, err := user.Current()
	if err != nil {
		if DEBUG {
			log.Fatal(err)
		} else {
			log.Fatal("Can't get user home directory path")
		}
	}
	newConfiguration := Configuration{
		Home:      usr.HomeDir,
		Directory: DirectoryName,
		Project:   ProjectDirectoryName,
	}
	newPathDirectory := newConfiguration.getDirectoryPath()
	if _, err := os.Stat(newPathDirectory); os.IsNotExist(err) {
		_ = os.Mkdir(newPathDirectory, os.ModePerm)
	}
	newPathProject := newConfiguration.GetProjectPath()
	if _, err := os.Stat(newPathProject); os.IsNotExist(err) {
		_ = os.Mkdir(newPathProject, os.ModePerm)
	}
	return newConfiguration
}

func (config Configuration) getDirectoryPath() string {
	return config.Home + "/" + config.Directory
}

func (config Configuration) GetProjectPath() string {
	return config.getDirectoryPath() + "/" + config.Project
}

func (config Configuration) GetProjectConfigFilePath() string {
	return config.GetProjectPath() + "/" + ProjectFileName
}
