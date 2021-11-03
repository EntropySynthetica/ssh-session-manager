package main

import (
	"encoding/json"
	"flag"
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

type Hosts struct {
	Hosts struct {
		Default []Host `json:"default"`
	} `json:"hosts"`
}

func main() {
	homedir, _ := os.UserHomeDir()
	var configFile string

	// Parse Command Line Flags
	newFile := flag.Bool("new", false, "Create a new hosts file")
	flag.StringVar(&configFile, "c", homedir+"/.config/ssm/hosts.json", "specify path of config file")
	flag.Parse()

	// If the new file flag is selected create a new hosts file.
	if *newFile {
		host1 := Host{}
		host1.Name = "localhost"
		host1.Hostname = "127.0.0.1"
		host1.User = "Username"

		host2 := Host{}
		host2.Name = "Router"
		host2.Hostname = "192.168.0.254"
		host2.User = "admin"

		jsonFile := Hosts{}
		jsonFile.Hosts.Default = append(jsonFile.Hosts.Default, host1)
		jsonFile.Hosts.Default = append(jsonFile.Hosts.Default, host2)

		file, _ := json.MarshalIndent(jsonFile, "", " ")
		_ = ioutil.WriteFile(configFile, file, 0644)
		return
	}

	// Open the hosts file
	jsonFile, err := os.Open(configFile)

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

	if jsonResults["hosts"] == nil {
		fmt.Println("Error, no host groups found in hosts.json")
		return
	}

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
