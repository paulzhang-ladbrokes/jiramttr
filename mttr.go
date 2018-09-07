package jiramttr

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

// IssueFields is selected fields of a jira issue. We are only interested in created and updated dates.
type IssueFields struct {
	Created  string   `json:"created"`
	Finished string   `json:"resolutiondate"`
	Labels   []string `json:"labels"`
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

var responseSecondsPerTeam = make(map[string]float64)

var issuesPerTeam = make(map[string]int32)

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
func GetMTTR(url string) (map[string]float64, error) {
	issues, err := getIssues(url + "&maxResults=500")
	if err != nil {
		return responseSecondsPerTeam, err
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

func getOwnersByIssue(issue Issue) []string {
	var owners []string
	for _, label := range issue.Fields.Labels {
		IssueOwners := GetOwners(label)
		if len(IssueOwners) == 0 {
			continue
		}
		owners = append(owners, IssueOwners...)
	}

	return owners
}

func parseResponse(issues []Issue) (map[string]float64, error) {
	var totalSeconds float64
	count := 0

	for _, issue := range issues {
		finished, err := dateparse.ParseLocal(issue.Fields.Finished)
		if err != nil {
			return responseSecondsPerTeam, err
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
			return responseSecondsPerTeam, err
		}

		responseTime := finished.Sub(created).Seconds()
		responseDay := responseTime / 3600.0 / 24.0

		owners := getOwnersByIssue(issue)

		if len(owners) == 0 {
			fmt.Printf("%s [%s] does not have an owner!\n",
				issue.Key,
				strings.Join(issue.Fields.Labels, ","),
			)
			continue
		}

		for _, team := range owners {
			_, ok := responseSecondsPerTeam[team]
			if !ok {
				responseSecondsPerTeam[team] = 0
				issuesPerTeam[team] = 0
			}
			responseSecondsPerTeam[team] += responseTime
			issuesPerTeam[team]++
		}

		fmt.Printf("%s took %.1f days, (%d-%02d-%02d - %d-%02d-%02d). [%s] has owners %s\n",
			issue.Key,
			responseDay,
			created.Year(),
			created.Month(),
			created.Day(),
			finished.Year(),
			finished.Month(),
			finished.Day(),
			strings.Join(issue.Fields.Labels, ","),
			strings.Join(owners, ","),
		)

		totalSeconds += responseTime
	}

	if count == 0 {
		return responseSecondsPerTeam, errors.New("No jira ticket is found")
	}

	for team, responseSeconds := range responseSecondsPerTeam {
		issueCount, ok := issuesPerTeam[team]
		if !ok {
			fmt.Println("I cannot believe it!")
			responseSecondsPerTeam[team] = 0
			continue
		}
		responseSecondsPerTeam[team] = responseSeconds / float64(issueCount)
	}

	responseSecondsPerTeam["__total__"] = totalSeconds / float64(count)

	return responseSecondsPerTeam, nil
}
