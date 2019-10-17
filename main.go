package main

import (
	"bufio"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/supanadit/devops-factory/system"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/src-d/go-git.v4"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/alexflint/go-arg"
	"github.com/ozgio/strutil"
	"github.com/supanadit/devops-factory/model"
)

type args struct {
	Pn string `arg:"separate" help:"New Project"`
	Pe string `arg:"separate" help:"New Project From Existing Repository"`
	Pr string `arg:"separate" help:"Remove Project"`
	Pl bool   `arg:"separate" help:"Project List"`
	Pu string `arg:"separate" help:"Project Git Update"`
	Kn string `arg:"separate" help:"New SSH Keyring"`
	Kr string `arg:"separate" help:"Remove SSH Keyring"`
	Kc string `arg:"separate" help:"Connect to SSH"`
	Kl bool   `arg:"separate" help:"List SSH Keyring"`
}

func (args) Version() string {
	return "DevOps Factory 0.0.4 Alpha"
}

func main() {
	var args args
	arg.MustParse(&args)
	cfg := model.LoadDefaultConfiguration()

	if args.Pn == "" && args.Pe == "" && args.Pr == "" && !args.Pl && args.Pu == "" && args.Kn == "" && args.Kr == "" && args.Kc == "" && !args.Kl {
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

		urlGitConversion := strings.TrimSuffix(urlGit, "\n")
		fmt.Printf("Cloning %s \n", urlGitConversion)
		_, err := git.PlainClone(project.Path, false, &git.CloneOptions{
			URL:      urlGitConversion,
			Progress: os.Stdout,
		})
		if err != nil {
			if model.DEBUG {
				log.Print(err)
			} else {
				fmt.Printf("Make sure URL Repository is correct \n")
			}
			_ = os.RemoveAll(project.Path)
			continueProcess = false
		}
		if continueProcess {
			continueProcess = project.Save(cfg)
		}
	}

	if args.Pe != "" {
		existingProject, _ := filepath.Abs(args.Pe)
		isExist := true
		if _, err := os.Stat(existingProject); os.IsNotExist(err) {
			isExist = false
		}
		if isExist {
			var project model.Project
			alias := strutil.Slugify(filepath.Base(existingProject))
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Project Name : ")
			projectName, _ := reader.ReadString('\n')
			project.ProjectName = strings.TrimSuffix(projectName, "\n")
			project.Alias = alias
			project.Path = existingProject
			r, err := git.PlainOpen(project.Path)
			// @TODO: Simplify the code
			if err != nil {
				if model.DEBUG {
					log.Print(err)
				} else {
					fmt.Printf("Cannot check git repository \n")
				}
			}
			if r != nil {
				err = r.Storer.PackRefs()
				if err != nil {
					if model.DEBUG {
						log.Print(err)
					} else {
						fmt.Printf("Git Repository doesn't exist \n")
					}
				} else {
					var remotes []*git.Remote
					remotes, err = r.Remotes()
					if err != nil {
						if model.DEBUG {
							log.Print(err)
						} else {
							fmt.Printf("Cannot check remote repository \n")
						}
					} else {
						if len(remotes) == 0 {
							fmt.Println("No Remote Repository exist")
						} else {
							project.Save(cfg)
						}
					}
				}
			}
		} else {
			fmt.Printf("Path for %s is not exist \n", existingProject)
		}
	}

	if args.Pr != "" {
		var project model.Project
		exist := false
		project.Alias = args.Pr
		exist, project = project.ExistByAlias(cfg)
		if exist {
			project.Remove(cfg)
		} else {
			fmt.Printf("Project with alias %s is not exist \n", args.Pr)
		}
	}

	if args.Pl {
		projectConfiguration := model.GetAllProjectConfiguration(cfg)
		var data [][]string
		for i, element := range projectConfiguration.Project {
			newData := []string{
				strconv.Itoa(i + 1),
				element.ProjectName,
				element.Alias,
				element.UrlRepository(cfg),
			}
			data = append(data, newData)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"No", "Project Name", "Alias", "Repository URL"})
		table.AppendBulk(data)
		table.Render()
	}

	if args.Pu != "" {
		var project model.Project
		exist := false
		project.Alias = args.Pu
		exist, project = project.ExistByAlias(cfg)
		if exist {
			var repository *git.Repository
			var exist bool
			repository, exist = project.GitRepository(cfg)
			if exist {
				var workTree *git.Worktree
				var err error
				workTree, err = repository.Worktree()
				if err == nil {
					err = workTree.Pull(&git.PullOptions{
						RemoteName: "origin",
						Progress:   os.Stdout,
					})
					if err != nil {
						if model.DEBUG {
							fmt.Println(err)
						} else {
							fmt.Println("Cannot Update Repository")
						}
					}
				} else {
					fmt.Println("Cannot getting Work Tree from the Repository")
				}
			} else {
				fmt.Println("Make sure repository exist")
			}
		} else {
			fmt.Printf("Project with alias %s is not exist \n", args.Pr)
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
					fmt.Println(err)
				} else {
					fmt.Println("Error while setup port")
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
					fmt.Println(err)
				} else {
					fmt.Println("Error while setup username")
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
					fmt.Println(err)
				} else {
					fmt.Println("Error while setup password")
				}
				continueProcess = false
			}
			// Line Break
			fmt.Println("")
			if continueProcess {
				keyringModel.Password = strings.TrimSuffix(string(bytePassword), "\n")
				if keyringModel.Exist(cfg) {
					fmt.Printf("Keyring SSH for %s with username %s is exist \n", keyringModel.Host, keyringModel.Username)
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
			fmt.Printf("Success Delete %s with username %s \n", keyring.Host, keyring.Username)
		} else {
			fmt.Println("Please specified keyring to delete eg. root@123.123.132.123")
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
					fmt.Println(err)
				} else {
					fmt.Println("Make sure username, password and port is correct")
				}
				continueProcess = false
			}
			if continueProcess {
				defer client.Close()
				if err := client.Terminal(nil).Start(); err != nil {
					if model.DEBUG {
						fmt.Println(err)
					} else {
						fmt.Println("Cannot open interactive shell")
					}
				}
			}
		} else {
			fmt.Println("Please specified keyring eg. root@123.123.132.123")
		}
	}

	if args.Kl {
		keyringList := model.GetAllKeyringConfiguration(cfg)
		var data [][]string
		for i, element := range keyringList.Keyring {
			newData := []string{
				strconv.Itoa(i + 1),
				element.Username,
				element.Host,
				element.Port,
			}
			data = append(data, newData)
		}
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"No", "Username", "Host", "Port"})
		table.AppendBulk(data)
		table.Render()
	}
}
