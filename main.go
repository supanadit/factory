package main

import (
	"bufio"
	"fmt"
	"github.com/supanadit/devops-factory/system"
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
	Kc string `arg:"separate" help:"Connect to SSH"`
}

func (args) Version() string {
	return "DevOps Factory 0.0.3 Beta"
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
		continueProcess := true
		var keyringModel = model.GetKeyringFromString(args.Kn)
		if keyringModel.Port == "" {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Port ( Left blank for default 22 ) : ")
			port, err := reader.ReadString('\n')
			if err != nil {
				if model.DEBUG {
					fmt.Print(err)
				} else {
					fmt.Print("Error while setup port")
				}
				continueProcess = false
			} else {
				keyringModel.Port = strings.TrimSuffix(port, "\n")
			}
		}

		if keyringModel.Username == "" {
			reader := bufio.NewReader(os.Stdin)
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
		var keyring = model.GetKeyringFromString(args.Kr)
		if keyring.Username != "" && keyring.Host != "" {
			keyring.RemoveFromAll(cfg)
			fmt.Printf("Success Delete %s with username %s", keyring.Host, keyring.Username)
		} else {
			fmt.Print("Please specified keyring to delete eg. root@123.123.132.123")
		}
	}

	if args.Kc != "" {
		continueProcess := true
		var keyring = model.GetKeyringFromString(args.Kc)
		if keyring.Username != "" && keyring.Host != "" {
			keyring.Password = keyring.GetPasswordFromSystem()
			client, err := system.DialWithPasswd(keyring.GetHostPort(), keyring.Username, keyring.Password)
			if err != nil {
				if model.DEBUG {
					fmt.Print(err)
				} else {
					fmt.Print("Make sure username, password and port is correct")
				}
				continueProcess = false
			}
			if continueProcess {
				defer client.Close()
				if err := client.Terminal(nil).Start(); err != nil {
					if model.DEBUG {
						fmt.Print(err)
					} else {
						fmt.Print("Cannot open interactive shell")
					}
				}
			}
		} else {
			fmt.Print("Please specified keyring eg. root@123.123.132.123")
		}
	}
}
