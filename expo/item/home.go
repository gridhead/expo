package item

import (
	"time"
)

type RepoData struct {
	RootSrce     string
	NameSrce     string
	RootDest     string
	NameDest     string
	PasswordSrce string
	PasswordDest string
	UsernameSrce string
	UsernameDest string
}

type ProjData struct {
	Id          int
	Name        string
	Desc        string
	Link        string
	DateCreated time.Time
	DateUpdated time.Time
}
