package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/paulzhang-ladbrokes/jiramttr"
)

var url = flag.String("url", "", "jira ticket search api url (eg. https://jira.com/rest/api/2/search?jql=xxx)")
var month = flag.String("month", "", "yyyy-mm to limit jira tickets that are finished within (eg. 2018-08)")

func main() {
	flag.Parse()

	err := jiramttr.SetMonth(*month)
	if err != nil {
		log.Fatalln("Got error setting month", err)
	}

	err = jiramttr.ReadOwners()
	if err != nil {
		log.Fatalln("Failed to get owners data from owners.json file", err)
	}

	mttr, err := jiramttr.GetMTTR(*url)
	if err != nil {
		log.Fatalln("Got error getting MTTR", err)
	}

	for team, response := range mttr {
		if team == "__total__" {
			continue
		}
		fmt.Printf("In %s: %s's MTTR is %d seconds, %.1f days\n",
			*month, team, int32(response), (response / 3600.0 / 24.0),
		)
	}
	overallMTTR := mttr["__total__"]
	fmt.Printf("In %s: overall MTTR is %d seconds, %.1f days\n",
		*month, int32(overallMTTR), (overallMTTR / 3600.0 / 24.0),
	)
}
