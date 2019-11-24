package model

import (
	"fmt"
	"github.com/naoina/toml"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

// Constant Value
const DEBUG bool = true
const SystemName string = "DevOpsFactory" // Don't Use any Space
const DirectoryName string = "DevOpsFactory"
const ProjectDirectoryName string = "Project"

const ProjectFileName string = "projects.toml"            // Format Must Be TOML
const ConfigurationFileName string = "configuration.toml" // Format Must Be TOML
const KeyringFileName string = "keyring.toml"             // Format Must Be TOML

const ConfirmationDeleteForkedRepository = "I AM REALLY SURE TO DELETE ALL FORKED REPOSITORY"

// Default Configuration
type Configuration struct {
	Home   string
	Github Github
}

type ProjectConfiguration struct {
	Project []Project
}

type KeyringConfiguration struct {
	Keyring []Keyring
}

func DefaultConfiguration() Configuration {
	usr, err := user.Current()
	if err != nil {
		if DEBUG {
			log.Fatal(err)
		} else {
			log.Fatal("Can't get user home directory path")
		}
	}
	configuration := Configuration{
		Home: usr.HomeDir,
		Github: Github{
			ID:       -1,
			Name:     "",
			Token:    "",
			Username: "",
		},
	}
	return configuration
}

func LoadConfiguration() Configuration {
	configuration, _ := GetConfiguration()

	newPathDirectory := configuration.GetDirectoryPath()
	if _, err := os.Stat(newPathDirectory); os.IsNotExist(err) {
		_ = os.Mkdir(newPathDirectory, os.ModePerm)
	}
	newPathProject := configuration.GetProjectPath()
	if _, err := os.Stat(newPathProject); os.IsNotExist(err) {
		_ = os.Mkdir(newPathProject, os.ModePerm)
	}
	return configuration
}

func (configuration Configuration) SaveConfiguration() {
	dataToml, _ := toml.Marshal(&configuration)
	err := ioutil.WriteFile(configuration.GetConfigFilePath(), dataToml, 0644)
	if err != nil {
		if DEBUG {
			log.Println(err)
		} else {
			fmt.Println("Cannot save configuration")
		}
	}
}

func (configuration Configuration) GetDirectoryPath() string {
	return configuration.Home + "/" + DirectoryName
}

func (configuration Configuration) GetProjectPath() string {
	return configuration.GetDirectoryPath() + "/" + ProjectDirectoryName
}

func (configuration Configuration) GetProjectConfigFilePath() string {
	return configuration.GetProjectPath() + "/" + ProjectFileName
}

func (configuration Configuration) GetKeyringConfigFilePath() string {
	return configuration.GetDirectoryPath() + "/" + KeyringFileName
}

func (configuration Configuration) GetConfigFilePath() string {
	return configuration.GetDirectoryPath() + "/" + ConfigurationFileName
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

func GetPublicKey() *ssh.PublicKeys {
	var publicKey *ssh.PublicKeys
	sshPath := os.Getenv("HOME") + "/.ssh/id_rsa"
	sshKey, _ := ioutil.ReadFile(sshPath)
	publicKey, keyError := ssh.NewPublicKeys("git", []byte(sshKey), "")
	if keyError != nil {
		fmt.Println("Invalid SSH Key")
	}
	return publicKey
}

func GetConfiguration() (Configuration, error) {
	configuration := DefaultConfiguration()
	info, configError := os.Stat(configuration.GetConfigFilePath())
	fileConfigurationExist := true
	if os.IsNotExist(configError) {
		fileConfigurationExist = false
	} else {
		fileConfigurationExist = !info.IsDir()
	}

	if !fileConfigurationExist {
		configuration.SaveConfiguration()
	} else {
		fileConfig, configError := os.Open(configuration.GetConfigFilePath())
		if configError != nil {
			if DEBUG {
				panic(configError)
			}
		}
		if configError = toml.NewDecoder(fileConfig).Decode(&configuration); configError != nil {
			if DEBUG {
				panic(configError)
			}
		}
	}
	return configuration, configError
}
