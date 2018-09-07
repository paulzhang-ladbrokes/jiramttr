package jiramttr

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

var domainOwners map[string][]string

// ReadOwners reads owners.json file in and fills domainOwners
func ReadOwners() error {
	pwd, _ := os.Getwd()
	data, err := ioutil.ReadFile(pwd + "/owners.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &domainOwners)
	if err != nil {
		return err
	}
	return nil
}

// GetOwners takes a domain and return owners
func GetOwners(domain string) []string {
	owners, ok := domainOwners[domain]
	if ok {
		return owners
	}
	return owners
}
