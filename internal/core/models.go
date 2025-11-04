package core

import "time"

type Job struct {
	ID        string
	Action    string
	RepoOwner string
	RepoName  string
	JobName   string
	Status    string
	CreatedAt time.Time
}
