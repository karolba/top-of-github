package main

import "time"

type Repo struct {
	Id            int64     `json:"id" xorm:"pk notnull"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	GithubLink    string    `json:"html_url"`
	Homepage      string    `json:"homepage"`
	Description   string    `json:"description"`
	Language      string    `json:"language"`
	Stargazers    int64     `json:"stargazers_count"`
	Topics        []string  `json:"topics"`
	OpenIssues    int64     `json:"open_issues_count"`
	Archived      bool      `json:"archived"`
	CreatedAt     time.Time `json:"created_at"`
	RepoPushedAt  time.Time `json:"pushed_at"`
	RepoUpdatedAt time.Time `json:"updated_at"`
	Owner         struct {
		Login      string `json:"login" xorm:"'OwnerLogin'"`
		AvatarUrl  string `json:"avatar_url" xorm:"'OwnerAvatarUrl'"`
		GravatarId string `json:"gravatar_id" xorm:"'OwnerGravatarId'"`
		Type       string `json:"type" xorm:"'OwnerType'"`
	} `json:"owner" xorm:"extends"`
	License struct {
		SpdxId string `json:"spdx_id" xorm:"'LicenseSpdxId'"`
	} `json:"license" xorm:"extends"`

	LastFetchedFromGithubAt  time.Time `json:"-"`
	FirstFetchedFromGithubAt time.Time `json:"-"`
}

type State struct {
	Name  string `xorm:"pk notnull"`
	Value string
}
