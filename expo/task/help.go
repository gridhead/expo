package task

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gridhead/expo/expo/item"
	"github.com/tidwall/gjson"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func IsIssueTicketClosed(data string) bool {
	var rslt bool = false
	if data == "Closed" {
		rslt = true
	}
	return rslt
}

func ValidateStatusChoice(data string) error {
	var StatusApplicable = map[string]bool{"OPEN": true, "SHUT": true, "FULL": true}
	var ReturnData error = nil
	if !StatusApplicable[data] {
		ReturnData = errors.New("Invalid status choice. Must be 'OPEN', 'SHUT' OR 'FULL'.")
	}
	return ReturnData
}

func ValidateRanges(data string) (int, int, error) {
	part := strings.Split(data, "-")
	var rmin, rmax int
	var expt error = nil

	if len(part) != 2 {
		expt = errors.New("Invalid range format. Must be X-Y (eg. 10-50).")
	} else {
		rmin, emin := strconv.Atoi(part[0])
		if emin != nil {
			expt = errors.New("Invalid lower value. Must be an integer.")
		}

		rmax, emax := strconv.Atoi(part[1])
		if emax != nil {
			expt = errors.New("Invalid upper value. Must be an integer.")
		}

		if rmin >= rmax {
			expt = errors.New("Invalid range. Lower value must be lesser than upper value.")
		}
	}
	return rmin, rmax, expt
}

func ActionStatusChoice(data string) string {
	var status string = "open"
	switch data {
	case "OPEN":
		status = "open"
	case "SHUT":
		status = "closed"
	case "FULL":
		status = "all"
	default:
		status = "open"
	}
	return status
}

func ValidateChoice(data string) ([]int, error) {
	part := strings.Split(data, ",")
	var list []int
	var expt error = nil
	for _, pack := range part {
		if pack == "" {
			continue
		}
		numb, cver := strconv.Atoi(pack)
		if cver != nil {
			expt = cver
			break
		}
		var prez bool
		for _, qant := range list {
			if qant == numb {
				prez = true
			}
		}
		if !prez {
			list = append(list, numb)
		}
	}
	return list, expt
}

func HTTPPagureGetSupplicant(base string, prms url.Values, password string, want int) (data string, expt error) {
	link := fmt.Sprintf("%s?%s", base, prms.Encode())
	rqst, expt := http.NewRequest("GET", link, nil)
	rqst.Header.Set("Authorization", fmt.Sprintf("token %s", password))
	oper := &http.Client{Timeout: 60 * time.Second}
	resp, expt := oper.Do(rqst)

	if expt != nil || resp.StatusCode != want {
		slog.Log(nil, slog.LevelError, "Failed to retrieve issue tickets from namespace")
		if expt != nil {
			slog.Log(nil, slog.LevelError, fmt.Sprintf("Error occured. %s", expt.Error()))
		}
		if resp.StatusCode != want {
			slog.Log(nil, slog.LevelError, fmt.Sprintf("Error occured. %s", resp.Status))
		}
	}
	defer resp.Body.Close()

	body, expt := io.ReadAll(resp.Body)
	if expt != nil {
		slog.Log(nil, slog.LevelError, "Failed to retrieve issue tickets from namespace")
		slog.Log(nil, slog.LevelError, fmt.Sprintf("Error occured. %s", expt.Error()))
	}

	return string(body), expt
}

// TODO Delete this function once the fetching function has been primed
func TempReadFileJSON() string {
	data, expt := os.ReadFile("/home/fedohide-origin/projects/expo/temp.json")
	if expt != nil {
		slog.Log(nil, slog.LevelError, fmt.Sprintf("Error occured. %s", expt.Error()))
	}
	return string(data)
}

func HTTPForgejoPostSupplicant(base string, data string, password string, want int) (string, error) {
	var rslt string

	link := fmt.Sprintf("%s", base)

	rqst, expt := http.NewRequest("POST", link, bytes.NewBuffer([]byte(data)))
	if expt != nil {
		return rslt, expt
	}

	rqst.Header.Set("Content-Type", "application/json")
	rqst.Header.Set("Authorization", fmt.Sprintf("token %s", password))
	oper := &http.Client{Timeout: 60 * time.Second}
	resp, expt := oper.Do(rqst)

	if expt != nil || resp.StatusCode != want {
		if expt != nil {
			return rslt, expt
		}
		if resp.StatusCode != want {
			return rslt, errors.New(fmt.Sprintf("Error occured. %s", resp.Status))
		}
	}
	defer resp.Body.Close()

	body, expt := io.ReadAll(resp.Body)
	if expt != nil {
		slog.Log(nil, slog.LevelError, "Failed to retrieve issue tickets from namespace")
		slog.Log(nil, slog.LevelError, fmt.Sprintf("Error occured. %s", expt.Error()))
	}

	return string(body), expt
}

func VerifySrceProject(repodata *item.RepoData) (*item.ProjData, error) {
	var burl, data string
	var prms url.Values
	var expt error
	var projdata item.ProjData
	burl = fmt.Sprintf("https://%s/api/0/%s", repodata.RootSrce, repodata.NameSrce)
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordSrce, 200)
	if expt != nil {
		return nil, expt
	}
	rsltdict := gjson.Parse(data)
	projdata = item.ProjData{
		Id:          int(rsltdict.Get("id").Int()),
		Name:        rsltdict.Get("fullname").String(),
		Desc:        rsltdict.Get("description").String(),
		Link:        rsltdict.Get("full_url").String(),
		DateCreated: time.Unix(rsltdict.Get("date_created").Int(), 0),
		DateUpdated: time.Unix(rsltdict.Get("date_modified").Int(), 0),
	}
	return &projdata, nil
}

func VerifyDestProject(repodata *item.RepoData) (*item.ProjData, error) {
	var burl, data string
	var prms url.Values
	var expt error
	var projdata item.ProjData
	var date_created, date_updated time.Time
	burl = fmt.Sprintf("https://%s/api/v1/repos/%s", repodata.RootDest, repodata.NameDest)
	data, expt = HTTPPagureGetSupplicant(burl, prms, repodata.PasswordDest, 200)
	if expt != nil {
		return nil, expt
	}

	rsltdict := gjson.Parse(data)

	date_created, expt = time.Parse(time.RFC3339, rsltdict.Get("created_at").String())
	if expt != nil {
		return nil, expt
	}

	date_updated, expt = time.Parse(time.RFC3339, rsltdict.Get("updated_at").String())
	if expt != nil {
		return nil, expt
	}

	projdata = item.ProjData{
		Id:          int(rsltdict.Get("id").Int()),
		Name:        rsltdict.Get("full_name").String(),
		Desc:        rsltdict.Get("description").String(),
		Link:        rsltdict.Get("html_url").String(),
		DateCreated: date_created,
		DateUpdated: date_updated,
	}
	return &projdata, nil
}

func VerifyProjects(repodata *item.RepoData) (bool, error) {
	var rslt bool
	var expt error

	slog.Log(nil, slog.LevelWarn, "○ Verifying source namespace...")
	srceproj, expt := VerifySrceProject(repodata)
	if expt != nil {
		rslt, expt = false, errors.New(fmt.Sprintf("Source namespace could not be verified. %s", expt.Error()))
	} else {
		rslt, expt = true, nil
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ ID.          %s", srceproj.Id))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Name.        %s", srceproj.Name))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Description. %s", srceproj.Desc))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Created on.  %s", srceproj.DateCreated.Format("Mon Jan 2 15:04:05 2006 UTC")))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Updated on.  %s", srceproj.DateUpdated.Format("Mon Jan 2 15:04:05 2006 UTC")))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✓ Source namespace verified"))
	}

	slog.Log(nil, slog.LevelWarn, "○ Verifying destination namespace...")
	destproj, expt := VerifyDestProject(repodata)
	if expt != nil {
		rslt, expt = false, errors.New(fmt.Sprintf("Destination namespace could not be verified. %s", expt.Error()))
	} else {
		rslt, expt = true, nil
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ ID.          %s", destproj.Id))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Name.        %s", destproj.Name))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Description. %s", destproj.Desc))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Created on.  %s", destproj.DateCreated.Format("Mon Jan 2 15:04:05 2006 UTC")))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("○ Updated on.  %s", destproj.DateUpdated.Format("Mon Jan 2 15:04:05 2006 UTC")))
		slog.Log(nil, slog.LevelInfo, fmt.Sprintf("✓ Destination namespace verified"))
	}

	return rslt, expt
}
