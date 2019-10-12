package model

import (
	"github.com/naoina/toml"
	"log"
	"os"
	"os/user"
)

// Constant Value
const DEBUG bool = false
const SystemName string = "DevOpsFactory" // Don't Use any Space
const DirectoryName string = "DevOpsFactory"
const ProjectDirectoryName string = "Project"
const ProjectFileName string = "projects.toml" // Format Must Be TOML
const KeyringFileName string = "keyring.toml"  // Format Must Be TOML

// Default Configuration
type Configuration struct {
	Home      string
	Directory string
	Project   string
}

type ProjectConfiguration struct {
	Project []Project
}

type KeyringConfiguration struct {
	Keyring []Keyring
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
	newPathDirectory := newConfiguration.GetDirectoryPath()
	if _, err := os.Stat(newPathDirectory); os.IsNotExist(err) {
		_ = os.Mkdir(newPathDirectory, os.ModePerm)
	}
	newPathProject := newConfiguration.GetProjectPath()
	if _, err := os.Stat(newPathProject); os.IsNotExist(err) {
		_ = os.Mkdir(newPathProject, os.ModePerm)
	}
	return newConfiguration
}

func (config Configuration) GetDirectoryPath() string {
	return config.Home + "/" + config.Directory
}

func (config Configuration) GetProjectPath() string {
	return config.GetDirectoryPath() + "/" + config.Project
}

func (config Configuration) GetProjectConfigFilePath() string {
	return config.GetProjectPath() + "/" + ProjectFileName
}

func (config Configuration) GetKeyringConfigFilePath() string {
	return config.GetDirectoryPath() + "/" + KeyringFileName
}

func GetAllProjectConfiguration(configuration Configuration) ProjectConfiguration {
	var projectConfiguration ProjectConfiguration
	if _, err := os.Stat(configuration.GetProjectConfigFilePath()); err == nil {
		fileConfig, fileError := os.Open(configuration.GetProjectConfigFilePath())
		if fileError != nil {
			panic(fileError)
		}
		if fileError = toml.NewDecoder(fileConfig).Decode(&projectConfiguration); fileError != nil {
			panic(fileError)
		}
	}
	return projectConfiguration
}

func GetAllKeyringConfiguration(configuration Configuration) KeyringConfiguration {
	var keyringConfiguration KeyringConfiguration
	if _, err := os.Stat(configuration.GetKeyringConfigFilePath()); err == nil {
		fileConfig, fileError := os.Open(configuration.GetKeyringConfigFilePath())
		if fileError != nil {
			panic(fileError)
		}
		if fileError = toml.NewDecoder(fileConfig).Decode(&keyringConfiguration); fileError != nil {
			panic(fileError)
		}
	}
	return keyringConfiguration
}
