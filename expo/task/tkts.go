package task

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gridhead/expo/expo/base"
	"github.com/gridhead/expo/expo/item"
	"github.com/tidwall/gjson"
	"log/slog"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func FetchTransferQuantity(repodata *item.RepoData, tktstask *item.TktsTaskData) (bool, error) {
	var burl, data string
	var prms url.Values
	var expt error
	var quantity int
	var rsltdict gjson.Result

	burl = fmt.Sprintf("https://%s/api/0/%s/issues", repodata.RootSrce, repodata.NameSrce)
	prms = url.Values{"status": {"all"}, "per_page": {strconv.Itoa(tktstask.PerPageQuantity)}, "page": {"1"}}
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		return false, expt
	}
	rsltdict = gjson.Get(data, "pagination")
	tktstask.PageQuantity = int(rsltdict.Get("pages").Int())

	prms = url.Values{"status": {"all"}, "per_page": {strconv.Itoa(tktstask.PerPageQuantity)}, "page": {strconv.Itoa(tktstask.PageQuantity)}}
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		return false, expt
	}

	rsltdict = gjson.Get(data, "total_issues")
	tktstask.IssueTicketQuantity = tktstask.PerPageQuantity*(tktstask.PageQuantity-1) + int(rsltdict.Int())

	burl = fmt.Sprintf("https://%s/api/0/%s/tags", repodata.RootSrce, repodata.NameSrce)
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		return false, expt
	}

	rsltdict = gjson.Get(data, "total_tags")
	tktstask.LabelsQuantity = int(rsltdict.Int())

	slog.Log(nil, slog.LevelWarn, fmt.Sprintf("✓ Found %d issue ticket(s) and %d issue label(s) across %d page(s).", tktstask.IssueTicketQuantity, tktstask.LabelsQuantity, tktstask.PageQuantity))

	if tktstask.WithLabels {
		_, expt = FetchLabelsInfo(repodata, tktstask)
		if expt != nil {
			return false, expt
		}
	}

	for indx := 1; indx < tktstask.PageQuantity+1; indx++ {
		slog.Log(nil, slog.LevelWarn, fmt.Sprintf("○ Fetching issue ticket(s) from Page #%d", indx))
		FetchIssueTicketsFromPage(repodata, tktstask, indx, &quantity)
	}
	slog.Log(nil, slog.LevelWarn, fmt.Sprintf("✓ Migrated %d issue ticket(s) out of %d issue ticket(s) successfully", quantity, tktstask.IssueTicketQuantity))

	return true, nil
}

func FetchLabelsInfo(repodata *item.RepoData, tktstask *item.TktsTaskData) (bool, error) {
	var burl, data string
	var prms url.Values
	var expt error
	var quantity int
	var rsltdict gjson.Result
	var list []string
	var waittags sync.WaitGroup

	burl = fmt.Sprintf("https://%s/api/0/%s/tags", repodata.RootSrce, repodata.NameSrce)
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		return false, expt
	}

	rsltdict = gjson.Get(data, "tags")
	for _, word := range rsltdict.Array() {
		list = append(list, word.String())
	}

	tktstask.LabelMap = make(map[string]int)
	for _, word := range list {
		waittags.Add(1)
		go MoveLabelsOver(repodata, tktstask, &quantity, &word, &waittags)
	}
	waittags.Wait()

	if quantity != tktstask.LabelsQuantity {
		return false, errors.New("Fetching labels failed")
	}

	slog.Log(nil, slog.LevelWarn, fmt.Sprintf("✓ Migrated %d issue label(s) out of %d issue label(s) successfully", quantity, tktstask.LabelsQuantity))

	return true, nil
}

func MoveLabelsOver(repodata *item.RepoData, tktstask *item.TktsTaskData, quantity *int, word *string, waittags *sync.WaitGroup) {
	defer waittags.Done()

	var burl, data, htmltext string
	var prms url.Values
	var expt error
	var unit item.TagsData
	var body item.TagsMakeBody
	var dict []byte
	var htmldict gjson.Result
	var tagsiden int
	var done bool

	burl = fmt.Sprintf("https://%s/api/0/%s/tag/%s", repodata.RootSrce, repodata.NameSrce, *word)
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ [%s] Issue label migration failed. %s", *word, expt.Error()))
		return
	}

	unit = item.TagsData{
		Name: gjson.Get(data, "tag").String(),
		Tint: gjson.Get(data, "tag_color").String(),
		Desc: gjson.Get(data, "tag_description").String(),
	}

	burl = fmt.Sprintf("https://%s/api/v1/repos/%s/labels", repodata.RootDest, repodata.NameDest)

	body = item.TagsMakeBody{
		Name:        unit.Name,
		Color:       unit.Tint,
		Description: unit.Desc,
		Exclusive:   false,
		IsArchived:  false,
	}

	dict, expt = json.Marshal(body)
	if expt != nil {
		slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ [%s] Issue label migration failed. %s", *word, expt.Error()))
		return
	}

	for indx := 0; indx < tktstask.Retries; indx++ {
		slog.Log(nil, slog.LevelDebug, fmt.Sprintf("○ [%s] Migrating issue label - Attempt %d of %d", *word, indx+1, tktstask.Retries))
		rslt, expt := HTTPForgejoPostSupplicant(burl, string(dict), repodata.PasswordDest, 201)
		if expt == nil {
			htmldict = gjson.Parse(rslt)
			htmltext = htmldict.Get("url").String()
			tagsiden = int(htmldict.Get("id").Int())
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✓ [%s] The issue label has been moved to %s", *word, htmltext))
			done = true
			break
		} else {
			slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ [%s] Issue label migration failed. %s", *word, expt.Error()))
		}
	}

	if done {
		tktstask.LabelMap[*word] = tagsiden
		*quantity = *quantity + 1
	}
}

func FetchIssueTicketsFromPage(repodata *item.RepoData, tktstask *item.TktsTaskData, indx int, quantity *int) {
	burl := fmt.Sprintf("https://%s/api/0/%s/issues", repodata.RootSrce, repodata.NameSrce)
	prms := url.Values{"status": {"all"}, "per_page": {strconv.Itoa(tktstask.PerPageQuantity)}, "page": {strconv.Itoa(indx)}}
	dump, expt := HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		slog.Log(nil, slog.LevelError, fmt.Sprintf("Error occured. %s", expt.Error()))
	}

	var wait sync.WaitGroup

	// data := string(TempReadFileJSON())
	data := gjson.Get(dump, "issues")

	for _, rootdict := range data.Array() {
		rsltdict := gjson.Parse(rootdict.String())
		var asgnobjc item.PersonData
		asgntext := rsltdict.Get("assignee").String()
		if asgntext != "" {
			asgndict := gjson.Parse(asgntext)
			asgnobjc = item.PersonData{
				FullUrl:  asgndict.Get("full_url").String(),
				FullName: asgndict.Get("fullname").String(),
				Name:     asgndict.Get("name").String(),
				UrlPath:  asgndict.Get("url_path").String(),
			}
		}
		var userobjc item.PersonData
		usertext := rsltdict.Get("user").String()
		if usertext != "" {
			userdict := gjson.Parse(usertext)
			userobjc = item.PersonData{
				FullUrl:  userdict.Get("full_url").String(),
				FullName: userdict.Get("fullname").String(),
				Name:     userdict.Get("name").String(),
				UrlPath:  userdict.Get("url_path").String(),
			}
		}
		var chatobjc []item.CommentData
		chattext := rsltdict.Get("comments").String()
		if chattext != "" {
			chatdict := gjson.Parse(chattext)
			for _, chatitem := range chatdict.Array() {
				srcetext := chatitem.Get("user").String()
				srcedict := gjson.Parse(srcetext)
				chatobjc = append(chatobjc, item.CommentData{
					Id:          int(chatitem.Get("id").Int()),
					Comment:     chatitem.Get("comment").String(),
					DateCreated: time.Unix(chatitem.Get("date_created").Int(), 0),
					User: item.PersonData{
						FullUrl:  srcedict.Get("full_url").String(),
						FullName: srcedict.Get("fullname").String(),
						Name:     srcedict.Get("name").String(),
						UrlPath:  srcedict.Get("url_path").String(),
					},
				})
			}
		}
		var tagsobjc []string
		tagstext := rsltdict.Get("tags").String()
		if tagstext != "" {
			tagsdict := gjson.Parse(tagstext)
			for _, tags := range tagsdict.Array() {
				tagsobjc = append(tagsobjc, tags.String())
			}
		}
		issuesObject := item.IssueTicketData{
			Title:       rsltdict.Get("title").String(),
			Id:          int(rsltdict.Get("id").Int()),
			Assignee:    asgnobjc,
			User:        userobjc,
			Content:     rsltdict.Get("content").String(),
			DateCreated: time.Unix(rsltdict.Get("date_created").Int(), 0),
			FullUrl:     rsltdict.Get("full_url").String(),
			Private:     rsltdict.Get("private").Bool(),
			Closed:      IsIssueTicketClosed(rsltdict.Get("status").String()),
			Tags:        tagsobjc,
			Comments:    chatobjc,
		}
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("▶ [#%d] %s by %s (%s) with %d comment(s)", issuesObject.Id, issuesObject.Title, issuesObject.User.FullName, issuesObject.User.Name, len(issuesObject.Comments)))
		wait.Add(1)
		go CreateIssueTicket(repodata, tktstask, &issuesObject, quantity, &wait)
	}
	wait.Wait()
}

func CreateIssueTicket(repodata *item.RepoData, tktstask *item.TktsTaskData, issuobjc *item.IssueTicketData, quantity *int, wait *sync.WaitGroup) {
	defer wait.Done()

	var htmldict gjson.Result
	var htmltext string
	var htmliden, chatnumb int
	var expt error
	var dict []byte
	var waitchat sync.WaitGroup
	var tgidlist []int
	//var work bool

	for _, unit := range issuobjc.Tags {
		for jndx, iden := range tktstask.LabelMap {
			if unit == jndx {
				tgidlist = append(tgidlist, iden)
			}
		}
	}

	data := item.TktsMakeBody{
		Title:  fmt.Sprintf(base.Headtemp, issuobjc.Id, issuobjc.Title),
		Body:   fmt.Sprintf(base.Bodytemp, issuobjc.Content, issuobjc.FullUrl, repodata.NameSrce, repodata.RootSrce, repodata.NameSrce, issuobjc.User.FullName, issuobjc.User.FullUrl, issuobjc.DateCreated.Format("Mon Jan 2 15:04:05 2006 UTC")),
		Labels: tgidlist,
	}

	if tktstask.WithStatus {
		data.Closed = issuobjc.Closed
	} else {
		data.Closed = false
	}

	dict, expt = json.Marshal(data)

	if expt != nil {
		slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ [#%d] Issue ticket migration failed. %s", issuobjc.Id, expt.Error()))
		return
	}

	burl := fmt.Sprintf("https://%s/api/v1/repos/%s/issues", repodata.RootDest, repodata.NameDest)
	for indx := 0; indx < tktstask.Retries; indx++ {
		slog.Log(nil, slog.LevelDebug, fmt.Sprintf("○ [#%d] Migrating issue ticket - Attempt %d of %d", issuobjc.Id, indx+1, tktstask.Retries))
		rslt, expt := HTTPForgejoPostSupplicant(burl, string(dict), repodata.PasswordDest, 201)
		if expt == nil {
			htmldict = gjson.Parse(rslt)
			htmltext = htmldict.Get("html_url").String()
			htmliden = int(htmldict.Get("number").Int())
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✓ [#%d] The issue ticket has been moved to %s", issuobjc.Id, htmltext))
			break
		} else {
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✗ [#%d] Issue ticket migration failed. %s", issuobjc.Id, expt.Error()))
		}
	}

	if htmliden == 0 {
		return
	}

	if tktstask.WithComments {
		for numb, unit := range issuobjc.Comments {
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("▷ [#%d] Comment %d of %d by %s (%s)", issuobjc.Id, numb+1, len(issuobjc.Comments), unit.User.FullName, unit.User.Name))
			waitchat.Add(1)
			go CreateIssueComment(repodata, &unit, issuobjc, &htmliden, &tktstask.Retries, &chatnumb, &waitchat)
		}
		waitchat.Wait()

		if chatnumb == len(issuobjc.Comments) {
			*quantity++
		}
	} else {
		*quantity++
	}
}

func CreateIssueComment(repodata *item.RepoData, unit *item.CommentData, issuobjc *item.IssueTicketData, htmliden *int, retries *int, chatnumb *int, waitchat *sync.WaitGroup) {
	defer waitchat.Done()

	var htmldict gjson.Result
	var htmltext string
	var expt error

	data := item.ChatMakeBody{
		Body: fmt.Sprintf(base.Chattemp, unit.Comment, issuobjc.FullUrl, unit.Id, unit.User.FullName, unit.User.FullUrl, issuobjc.FullUrl, repodata.NameSrce, repodata.RootSrce, repodata.NameSrce, unit.DateCreated.Format("Mon Jan 2 15:04:05 2006 UTC")),
	}
	dict, expt := json.Marshal(data)
	if expt != nil {
		slog.Log(nil, slog.LevelError, fmt.Sprintf("✗ [#%d] Issue comment migration failed. %s", issuobjc.Id, expt.Error()))
		return
	}

	burl := fmt.Sprintf("https://%s/api/v1/repos/%s/issues/%d/comments", repodata.RootDest, repodata.NameDest, *htmliden)

	for indx := 0; indx < *retries; indx++ {
		slog.Log(nil, slog.LevelDebug, fmt.Sprintf("○ [#%d] Migrating comment - Attempt %d of %d", issuobjc.Id, indx+1, *retries))
		rslt, expt := HTTPForgejoPostSupplicant(burl, string(dict), repodata.PasswordDest, 201)
		if expt == nil {
			htmldict = gjson.Parse(rslt)
			htmltext = htmldict.Get("html_url").String()
			*chatnumb = *chatnumb + 1
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✓ [#%d] The comment has been moved to %s", issuobjc.Id, htmltext))
			break
		} else {
			slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✗ [#%d] Issue comment migration failed. %s", issuobjc.Id, expt.Error()))
		}
	}
}
