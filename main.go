package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
)

type Host struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	User     string `json:"user"`
}

type Groups struct {
	Groupname string `json:"groupname"`
	Hosts     []Host `json:"hosts"`
}

type Hostfile struct {
	Groups []Groups `json:"groups"`
}

func NewFile(configFile string) {
	host1 := Host{}
	host1.Name = "localhost"
	host1.Hostname = "127.0.0.1"
	host1.User = "Username"

	host2 := Host{}
	host2.Name = "Router"
	host2.Hostname = "192.168.0.254"
	host2.User = "admin"

	host3 := Host{}
	host3.Name = "Linux1"
	host3.Hostname = "192.168.0.1"
	host3.User = "admin"

	group1 := Groups{}
	group1.Groupname = "Default"
	group1.Hosts = append(group1.Hosts, host1)
	group1.Hosts = append(group1.Hosts, host2)

	group2 := Groups{}
	group2.Groupname = "Workstations"
	group2.Hosts = append(group2.Hosts, host3)

	jsonFile := Hostfile{}
	jsonFile.Groups = append(jsonFile.Groups, group1)
	jsonFile.Groups = append(jsonFile.Groups, group2)

	file, _ := json.MarshalIndent(jsonFile, "", " ")
	_ = ioutil.WriteFile(configFile, file, 0644)
}

func AddGroup(jsonResults Hostfile, addGroup string, configFile string) {
	var newGroup Groups
	newGroup.Groupname = addGroup

	jsonResults.Groups = append(jsonResults.Groups, newGroup)

	file, _ := json.MarshalIndent(jsonResults, "", " ")
	_ = ioutil.WriteFile(configFile, file, 0644)

	fmt.Println("Adding new group: " + addGroup)
	return
}

func DeleteGroup(jsonResults Hostfile, delGroup string, configFile string) {
	var newJson Hostfile
	var foundGroup = false

	for i, group := range jsonResults.Groups {
		if group.Groupname == delGroup {
			foundGroup = true
		} else {
			newJson.Groups = append(newJson.Groups, jsonResults.Groups[i])
		}
	}

	if *&foundGroup {
		fmt.Println("Delete Group: " + delGroup)
		file, _ := json.MarshalIndent(newJson, "", " ")
		_ = ioutil.WriteFile(configFile, file, 0644)
	} else {
		fmt.Println("Error: Group " + delGroup + " was not found.")
	}
	return
}

func main() {
	homedir, _ := os.UserHomeDir()
	var configFile string
	var addGroup string
	var delGroup string

	// Parse Command Line Flags
	newFile := flag.Bool("new", false, "Create a new hosts file")
	flag.StringVar(&addGroup, "addgroup", "", "Add a new Group to the hosts file")
	flag.StringVar(&delGroup, "delgroup", "", "Delete a Group from the host file")
	flag.StringVar(&configFile, "c", homedir+"/.config/ssm/hosts.json", "specify path of config file")
	flag.Parse()

	// If the new file flag is selected create a new hosts file.
	if *newFile {
		NewFile(configFile)
		return
	}

	// Open the hosts file
	jsonFile, err := os.Open(configFile)

	if err != nil {
		fmt.Println(err)
		return
	}

	// Unmarshal the JSON from the file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var jsonResults Hostfile
	err = json.Unmarshal(byteValue, &jsonResults)

	if err != nil {
		fmt.Println("Error reading hosts.json ", err)
		return
	}

	jsonFile.Close()

	if addGroup != "" {
		AddGroup(jsonResults, addGroup, configFile)
		return
	}

	if delGroup != "" {
		DeleteGroup(jsonResults, delGroup, configFile)
		return
	}

	// Create emtpy arrays and dump the host into it.
	deviceNames := []string{}
	hostNames := []string{}
	users := []string{}
	groups := []string{}

	// Itterate thru the JSON and add host group names to an array.
	for _, group := range jsonResults.Groups {
		groups = append(groups, group.Groupname)
	}

	if groups == nil {
		fmt.Println("Error, no host groups found in hosts.json")
		return
	}

	groups = append(groups, "All")

	// Bring up the menu for the user to select the group to filter to.
	var selectGroup string
	groupPrompt := &survey.Select{
		Message: "Select a Device Group:",
		Options: groups,
	}
	survey.AskOne(groupPrompt, &selectGroup, survey.WithPageSize(10))

	// Itterate thru the JSON data and store a list of devices, hostsnames and users to arrays

	if selectGroup == "All" {
		// If the All group was selected we want to add every host
		for _, hosts := range jsonResults.Groups {
			for _, host := range hosts.Hosts {
				deviceNames = append(deviceNames, host.Name)
				hostNames = append(hostNames, host.Hostname)
				users = append(users, host.User)
			}
		}
	} else {
		// Otherwise just add the hosts for the group selected
		for i, group := range jsonResults.Groups {
			if group.Groupname == selectGroup {
				for _, host := range jsonResults.Groups[i].Hosts {
					deviceNames = append(deviceNames, host.Name)
					hostNames = append(hostNames, host.Hostname)
					users = append(users, host.User)
				}
			}
		}
	}

	if len(deviceNames) == 0 {
		fmt.Println("Error: No Hosts found in group " + selectGroup)
		return
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
