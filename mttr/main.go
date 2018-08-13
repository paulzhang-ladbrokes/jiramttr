package main

import (
	"flag"
	"log"

	"github.com/paulzhang-ladbrokes/jiramttr"
)

var url = flag.String("url", "", "jira ticket search api url (eg. https://jira.com/rest/api/2/search?jql=xxx)")

func main() {
	flag.Parse()

	mttr, err := jiramttr.GetMTTR(*url)
	if err != nil {
		log.Fatalln("Got error", err)
	}

	log.Printf("MTTR is %d seconds, %.1f days\n", int32(mttr), (mttr / 3600.0 / 24.0))
}
