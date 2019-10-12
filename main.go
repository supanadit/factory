package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/src-d/go-git.v4"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/naoina/toml"
	"github.com/ozgio/strutil"
	"github.com/supanadit/devops-factory/model"
)

type args struct {
	Pn string `arg:"separate" help:"New Project"`
	Kn string `arg:"separate" help:"New SSH Keyring"`
	Kr string `arg:"separate" help:"Remove SSH Keyring"`
}

func (args) Version() string {
	return "DevOps Factory 0.0.1 Beta"
}

func main() {
	var args args
	arg.MustParse(&args)
	cfg := model.LoadDefaultConfiguration()

	if args.Pn == "" && args.Kn == "" && args.Kr == "" {
		fmt.Println("Cross Platform Swiss Army Knife for DevOps")
	}

	if args.Pn != "" {
		continueProcess := true
		var project model.Project
		alias := strutil.Slugify(args.Pn)
		project.ProjectName = args.Pn
		project.Alias = alias
		project.Path = cfg.GetProjectPath() + "/" + project.Alias

		newProjectPath := project.Path
		if _, err := os.Stat(newProjectPath); os.IsNotExist(err) {
			_ = os.Mkdir(newProjectPath, os.ModePerm)
		}

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("URL Git Repository : ")
		urlGit, _ := reader.ReadString('\n')

		var gitModel model.Git
		gitModel.Url = strings.TrimSuffix(urlGit, "\n")
		gitModel.Path = project.Path
		project.Git = gitModel

		var allProject = model.GetAllProjectConfiguration(cfg)
		allProject.Project = append(allProject.Project, project)

		dataToml, _ := toml.Marshal(&allProject)
		_, err := git.PlainClone(gitModel.Path, false, &git.CloneOptions{
			URL:      gitModel.Url,
			Progress: os.Stdout,
		})
		if err != nil {
			if model.DEBUG {
				log.Print(err)
			} else {
				fmt.Printf("Make sure URL Repository is correct")
			}
			_ = os.RemoveAll(gitModel.Path)
			continueProcess = false
		}
		if continueProcess {
			err = ioutil.WriteFile(cfg.GetProjectConfigFilePath(), dataToml, 0644)
			if err != nil {
				if model.DEBUG {
					log.Print(err)
				} else {
					fmt.Printf("Cannot create configuration file for project")
				}
				continueProcess = false
			}
		}
	}

	if args.Kn != "" {
		userHost := strings.Split(args.Kn, "@")
		continueProcess := true
		var keyringModel model.Keyring
		reader := bufio.NewReader(os.Stdin)
		if len(userHost) > 1 {
			keyringModel.Username = userHost[0]
			keyringModel.Host = userHost[1]
		} else {
			fmt.Print("Username : ")
			name, err := reader.ReadString('\n')
			if err != nil {
				if model.DEBUG {
					fmt.Print(err)
				} else {
					fmt.Print("Error while setup username")
				}
				continueProcess = false
			} else {
				keyringModel.Username = strings.TrimSuffix(name, "\n")
				keyringModel.Host = args.Kn
			}
		}

		if continueProcess {
			fmt.Print("Password : ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				if model.DEBUG {
					fmt.Print(err)
				} else {
					fmt.Print("Error while setup password")
				}
				continueProcess = false
			}
			// Line Break
			fmt.Println("")
			if continueProcess {
				keyringModel.Password = strings.TrimSuffix(string(bytePassword), "\n")
				if keyringModel.Exist(cfg) {
					fmt.Printf("Keyring SSH for %s with username %s is exist", keyringModel.Host, keyringModel.Username)
				} else {
					keyringModel.SaveFull(cfg)
				}
			}
		}
	}

	if args.Kr != "" {
		userHost := strings.Split(args.Kr, "@")
		if len(userHost) > 1 {
			var keyring model.Keyring = model.Keyring{
				Host:     userHost[1],
				Username: userHost[0],
			}
			keyring.RemoveFromAll(cfg)
			fmt.Printf("Success Delete %s with username %s", keyring.Host, keyring.Username)
		} else {
			fmt.Print("Please specified keyring to delete eg. root@123.123.132.123")
		}
	}
}
