package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
	"github.com/mitchellh/mapstructure"
)

type Host struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	User     string `json:"user"`
}

func main() {

	// Open the hosts file
	homedir, _ := os.UserHomeDir()
	jsonFile, err := os.Open(homedir + "/.config/ssm/hosts.json")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer jsonFile.Close()

	// Unmarshal the JSON from the file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var jsonResults map[string]interface{}
	err = json.Unmarshal(byteValue, &jsonResults)

	if err != nil {
		fmt.Println("Error reading hosts.json ", err)
		return
	}

	// Create emtpy arrays and dump the host into it.
	deviceNames := []string{}
	hostNames := []string{}
	users := []string{}
	groups := []string{}

	// Itterate thru the JSON and add host group names to an array.
	groups = append(groups, "All")
	for key := range jsonResults["hosts"].(map[string]interface{}) {
		groups = append(groups, key)
	}

	// Bring up the menu for the user to select the group to filter to.
	var selectGroup string
	groupPrompt := &survey.Select{
		Message: "Select a Device Group:",
		Options: groups,
	}
	survey.AskOne(groupPrompt, &selectGroup, survey.WithPageSize(10))

	// Itterate thru the JSON data and store a list of devices, hostsnames and users to arrays
	for key, group := range jsonResults["hosts"].(map[string]interface{}) {
		// If the group is All then we want to match on every pass.
		if selectGroup == "All" {
			for _, hosts := range group.([]interface{}) {
				var host Host
				mapstructure.Decode(hosts, &host)
				deviceNames = append(deviceNames, host.Name)
				hostNames = append(hostNames, host.Hostname)
				users = append(users, host.User)
			}
			// Otherwise only add hosts from the selected group.
		} else if selectGroup == key {
			for _, hosts := range group.([]interface{}) {
				var host Host
				mapstructure.Decode(hosts, &host)
				deviceNames = append(deviceNames, host.Name)
				hostNames = append(hostNames, host.Hostname)
				users = append(users, host.User)
			}
		}
	}

	// Bring up the menu for the user to select the device to connect to.
	var selectDevice int
	prompt := &survey.Select{
		Message: "Select a device to connect to:",
		Options: deviceNames,
	}
	survey.AskOne(prompt, &selectDevice, survey.WithPageSize(10))

	// Run SSH and pass it the hostname of the device to connect to.
	fmt.Println("Connecting to " + deviceNames[selectDevice] + " - " + hostNames[selectDevice] + " as " + users[selectDevice])
	sshArgs := "ssh " + users[selectDevice] + "@" + hostNames[selectDevice]
	cmd := exec.Command("bash", "-c", sshArgs)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
