package task

import (
	"bytes"
	"errors"
	"fmt"
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
