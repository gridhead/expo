package item

import "time"

type PersonData struct {
	FullUrl  string
	FullName string
	Name     string
	UrlPath  string
}

type CommentData struct {
	Id          int
	Comment     string
	DateCreated time.Time
	User        PersonData
}

type IssueTicketData struct {
	Assignee    PersonData
	Comments    []CommentData
	Content     string
	DateCreated time.Time
	FullUrl     string
	Id          int
	Private     bool
	Closed      bool
	Tags        []string
	Title       string
	User        PersonData
}

type IssueTicketRanges struct {
	Min int
	Max int
}

type TktsTaskData struct {
	PageQuantity        int
	IssueTicketQuantity int
	PerPageQuantity     int
	LabelsQuantity      int
	Ranges              IssueTicketRanges
	Choice              []int
	Status              string
	WithComments        bool
	WithLabels          bool
	WithStatus          bool
	WithSecret          bool
	Retries             int
	LabelMap            map[string]int
}

type TktsMakeBody struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Closed bool   `json:"closed"`
	Labels []int  `json:"labels"`
}

type ChatMakeBody struct {
	Body string `json:"body"`
}

type TagsMakeBody struct {
	Name        string `json:"name"`
	Color       string `json:"color"`
	Description string `json:"description"`
	Exclusive   bool   `json:"exclusive"`
	IsArchived  bool   `json:"is_archived"`
}

type TagsData struct {
	Name string
	Tint string
	Desc string
}
