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

	mttr, err := jiramttr.GetMTTR(*url)
	if err != nil {
		log.Fatalln("Got error getting MTTR", err)
	}

	fmt.Printf("In %s: MTTR is %d seconds, %.1f days\n", *month, int32(mttr), (mttr / 3600.0 / 24.0))
}
