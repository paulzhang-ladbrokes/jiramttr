package jiramttr

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/araddon/dateparse"
)

// IssueFields test
type IssueFields struct {
	Created string `json:"created"`
	Updated string `json:"updated"`
}

// Issue test
type Issue struct {
	ID     string      `json:"id"`
	Fields IssueFields `json:"fields"`
}

// JiraResponse test
type JiraResponse struct {
	StartAt    int32   `json:"startAt"`
	MaxResults int32   `json:"maxResults"`
	Total      int32   `json:"total"`
	Issues     []Issue `json:"issues"`
}

var responseSeconds = make(map[string]float64)

// GetMTTR gets MTTR from jira tickets given an url
func GetMTTR(url string) (float64, error) {
	JiraResponse, err := getIssues(url)
	if err != nil {
		return 0.0, err
	}

	return parseResponse(JiraResponse)
}

func getIssues(url string) (*JiraResponse, error) {
	var jiraResponse *JiraResponse

	resp, err := http.Get(url)
	if err != nil {
		return jiraResponse, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return jiraResponse, err
	}

	err = json.Unmarshal(body, &jiraResponse)
	if err != nil {
		return nil, err
	}
	return jiraResponse, nil
}

func parseResponse(jiraResponse *JiraResponse) (float64, error) {
	var totalSeconds float64
	count := 0
	for _, issue := range jiraResponse.Issues {
		count++
		log.Printf("jira ticket (%s) is created at %s and updated at %s.\n", issue.ID, issue.Fields.Created, issue.Fields.Updated)

		created, err := dateparse.ParseLocal(issue.Fields.Created)
		if err != nil {
			return 0.0, err
		}

		finished, err := dateparse.ParseLocal(issue.Fields.Updated)
		if err != nil {
			return 0.0, err
		}

		id := issue.ID
		responseTime := finished.Sub(created).Seconds()
		responseSeconds[id] = responseTime
		totalSeconds += responseTime
	}

	if count == 0 {
		return 0.0, errors.New("No jira ticket is found")
	}

	return totalSeconds / float64(count), nil
}
