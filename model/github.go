package model

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Github struct {
	ID       int64
	Name     string
	Username string
	Token    string
}

func (configuration Configuration) GetGithubInformation() Github {
	return configuration.Github
}

func (configuration Configuration) SetToken(token string) {
	githubConfiguration := configuration.Github
	githubConfiguration.Token = token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Println(err)
	} else {
		githubConfiguration.Name = *user.Name
		githubConfiguration.ID = *user.ID
		githubConfiguration.Username = user.GetLogin()
	}
	configuration.Github = githubConfiguration
	configuration.SaveConfiguration()
}

func VerifyGithub(githubModel Github) bool {
	valid := false
	if githubModel.ID != -1 && githubModel.Name != "" && githubModel.Token != "" {
		valid = true
	} else {
		fmt.Println("Please verify your token or generate github token and save it with flag --gt")
	}
	return valid
}

func (githubModel Github) Client() github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubModel.Token},
	)
	tc := oauth2.NewClient(githubModel.Context(), ts)

	client := github.NewClient(tc)
	return *client
}

func (githubModel Github) Context() context.Context {
	return context.Background()
}
