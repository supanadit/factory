package model

type Project struct {
	ProjectName string
	Alias       string
	Git         Git
	Path        string
}
