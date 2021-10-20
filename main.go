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
	Name string `json:"name"`
	Ip   string `json:"ip"`
	User string `json:"user"`
}

func main() {

	// Open the hosts file
	jsonFile, err := os.Open("hosts.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	// Unmarshal the JSON from the file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var jsonResults map[string]interface{}
	json.Unmarshal(byteValue, &jsonResults)

	// Create emtpy arrays and dump the host into it.
	deviceNames := []string{}
	hostNames := []string{}
	users := []string{}

	// Itterate thru the JSON data and store a list of devices, hostsnames and users to the appove arrays
	for _, group := range jsonResults["hosts"].(map[string]interface{}) {
		for _, hosts := range group.([]interface{}) {
			var host Host
			mapstructure.Decode(hosts, &host)
			deviceNames = append(deviceNames, host.Name)
			hostNames = append(hostNames, host.Ip)
			users = append(users, host.User)
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
