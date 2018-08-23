package jiramttr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/araddon/dateparse"
)

// IssueFields is selected fields of a jira issue. We are only interested in created and updated dates.
type IssueFields struct {
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// Issue is a jira issue (or ticket). We are only interested in id and fields.
type Issue struct {
	ID     string      `json:"id"`
	Fields IssueFields `json:"fields"`
	Key    string      `json:"key"`
}

// JiraResponse is a jira response of a restful api request
type JiraResponse struct {
	StartAt    int32   `json:"startAt"`
	MaxResults int32   `json:"maxResults"`
	Total      int32   `json:"total"`
	Issues     []Issue `json:"issues"`
}

var responseSeconds = make(map[string]float64)

var month string

var startTime time.Time

var endTime time.Time

// SetMonth specifies a month to measure MTTR
func SetMonth(monthStr string) error {
	month = monthStr

	s, err := dateparse.ParseLocal(month)
	if err != nil {
		return err
	}

	startTime = s

	endTime = startTime.AddDate(0, 1, 0)

	return nil
}

// GetMTTR gets MTTR from jira issues given an url
func GetMTTR(url string) (float64, error) {
	issues, err := getIssues(url + "&maxResults=500")
	if err != nil {
		return 0.0, err
	}

	return parseResponse(issues)
}

func getIssues(url string) ([]Issue, error) {
	var jiraResponse *JiraResponse

	var issues []Issue

	resp, err := http.Get(url)
	if err != nil {
		return issues, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return issues, err
	}

	err = json.Unmarshal(body, &jiraResponse)
	if err != nil {
		return nil, err
	}

	return jiraResponse.Issues, nil
}

func parseResponse(issues []Issue) (float64, error) {
	var totalSeconds float64
	count := 0

	for _, issue := range issues {
		finished, err := dateparse.ParseLocal(issue.Fields.Updated)
		if err != nil {
			return 0.0, err
		}

		if finished.Before(startTime) {
			continue
		}

		if finished.After(endTime) {
			continue
		}

		count++

		created, err := dateparse.ParseLocal(issue.Fields.Created)
		if err != nil {
			return 0.0, err
		}

		id := issue.ID
		responseTime := finished.Sub(created).Seconds()
		responseDay := responseTime / 3600.0 / 24.0
		responseSeconds[id] = responseTime

		fmt.Printf("%s took %.1f days, (%d-%02d-%02d - %d-%02d-%02d).\n",
			issue.Key,
			responseDay,
			created.Year(),
			created.Month(),
			created.Day(),
			finished.Year(),
			finished.Month(),
			finished.Day(),
		)

		totalSeconds += responseTime
	}

	if count == 0 {
		return 0.0, errors.New("No jira ticket is found")
	}

	return totalSeconds / float64(count), nil
}
