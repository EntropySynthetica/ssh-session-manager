package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

const Version = "1.1.2"

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

	if _, err := os.Stat(configFile); err == nil {
		fmt.Println(configFile + " already exists!")
		confirm := false
		prompt := &survey.Confirm{
			Message: "Do you want to overwrite?",
		}
		survey.AskOne(prompt, &confirm)

		if confirm == true {
			file, _ := json.MarshalIndent(jsonFile, "", " ")
			_ = ioutil.WriteFile(configFile, file, 0644)
			fmt.Println("New hosts file created at " + configFile)
		}

	} else {
		// Check if config directory exists, and make it if it does not.
		homedir, _ := os.UserHomeDir()
		_, err := os.Stat(homedir + "/.config/ssm")
		if os.IsNotExist(err) {
			err := os.Mkdir(homedir+"/.config/ssm", 0755)
			if err != nil {
				log.Fatal(err)
			}
		}

		file, _ := json.MarshalIndent(jsonFile, "", " ")
		_ = ioutil.WriteFile(configFile, file, 0644)
		fmt.Println("New hosts file created at " + configFile)
	}
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

func AddHost(jsonResults Hostfile, groups []string, configFile string) {
	fmt.Println("Adding a host")

	var selectGroup string

	groupPrompt := &survey.Select{
		Message: "Select a Device Group:",
		Options: groups,
	}
	survey.AskOne(groupPrompt, &selectGroup, survey.WithPageSize(10))

	var newHost Host

	in := bufio.NewReader(os.Stdin)

	fmt.Println("Enter the details for the new host...")

	// Get the Name of the new host.
	fmt.Println("\nEnter Name: ")
	name, _ := in.ReadString('\n')
	if name == "\n" {
		fmt.Println("Error: Name can not be blank")
		return
	} else {
		newHost.Name = strings.TrimSuffix(name, "\n")
	}

	// Get the Hostname of the new host.
	fmt.Println("\nEnter Hostname as an IP or FQDN:")
	hostname, _ := in.ReadString('\n')
	if hostname == "\n" {
		fmt.Println("Error: Hostname can not be blank")
		return
	} else {
		newHost.Hostname = strings.TrimSuffix(hostname, "\n")
	}

	// Get the Username of the new host.
	fmt.Println("\nEnter username to log into host with:")
	user, _ := in.ReadString('\n')
	if user == "\n" {
		fmt.Println("Error: Username can not be blank")
		return
	} else {
		newHost.User = strings.TrimSuffix(user, "\n")
	}

	fmt.Println("\nYou have entered,")
	fmt.Println("Name: " + newHost.Name)
	fmt.Println("Hostname: " + newHost.Hostname)
	fmt.Println("Username: " + newHost.User)
	fmt.Println("Device Group: " + selectGroup)

	confirm := false
	prompt := &survey.Confirm{
		Message: "Is this correct?",
	}
	survey.AskOne(prompt, &confirm)

	if confirm == true {
		for i, group := range jsonResults.Groups {
			if group.Groupname == selectGroup {
				jsonResults.Groups[i].Hosts = append(jsonResults.Groups[i].Hosts, newHost)
			}
		}
		file, _ := json.MarshalIndent(jsonResults, "", " ")
		_ = ioutil.WriteFile(configFile, file, 0644)
	}
	return
}

func DelHost(jsonResults Hostfile, groups []string, configFile string) {
	fmt.Println("Deleting a Host")

	// Bring up the menu for the user to select the group to filter to.
	var selectGroup string
	groupPrompt := &survey.Select{
		Message: "Select a Device Group:",
		Options: groups,
	}
	survey.AskOne(groupPrompt, &selectGroup, survey.WithPageSize(10))

	deviceNames := []string{}
	var groupInt int

	for i, group := range jsonResults.Groups {
		if group.Groupname == selectGroup {
			groupInt = i
			for _, host := range jsonResults.Groups[i].Hosts {
				deviceNames = append(deviceNames, host.Name)
			}
		}
	}

	if len(deviceNames) == 0 {
		fmt.Println("Error: No Hosts found in group " + selectGroup)
		return
	}

	var selectDevice string
	prompt := &survey.Select{
		Message: "Select a device to connect to:",
		Options: deviceNames,
	}
	survey.AskOne(prompt, &selectDevice, survey.WithPageSize(10))

	var newGroup Groups

	for _, host := range jsonResults.Groups[groupInt].Hosts {
		newGroup.Groupname = jsonResults.Groups[groupInt].Groupname
		if host.Name != selectDevice {
			newGroup.Hosts = append(newGroup.Hosts, host)
		}
	}

	jsonResults.Groups[groupInt] = newGroup

	confirm := false
	delPrompt := &survey.Confirm{
		Message: "Are you sure you want to delete device " + selectDevice,
	}
	survey.AskOne(delPrompt, &confirm)

	if confirm == true {
		file, _ := json.MarshalIndent(jsonResults, "", " ")
		_ = ioutil.WriteFile(configFile, file, 0644)

		fmt.Println("Device "+selectDevice, " has been deleted from group "+jsonResults.Groups[groupInt].Groupname)
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
	addHost := flag.Bool("addhost", false, "Add a new Host to a group")
	delHost := flag.Bool("delhost", false, "Delete a Host")
	showVer := flag.Bool("v", false, "Show Version")
	flag.StringVar(&addGroup, "addgroup", "", "Add a new Group to the hosts file")
	flag.StringVar(&delGroup, "delgroup", "", "Delete a Group from the host file")
	flag.StringVar(&configFile, "c", homedir+"/.config/ssm/hosts.json", "specify path of config file")
	flag.Parse()

	// Print Version of app
	if *showVer {
		fmt.Println("SSH Session Manager Version: " + Version)
		return
	}

	// If the new file flag is selected create a new hosts file.
	if *newFile {
		NewFile(configFile)
		return
	}

	// Open the hosts file
	jsonFile, err := os.Open(configFile)

	if err != nil {
		fmt.Println(err)
		fmt.Println("You can create an example hosts file with the -new flag")
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

	// Runs if addgroup flag has something in it.
	if addGroup != "" {
		AddGroup(jsonResults, addGroup, configFile)
		return
	}

	// Runs if deletegroup flag has somethign in it.
	if delGroup != "" {
		DeleteGroup(jsonResults, delGroup, configFile)
		return
	}

	// Create emtpy arrays and dump the host into it.
	deviceNames := []string{}
	hostNames := []string{}
	users := []string{}
	groups := []string{}

	// Iterate through the JSON and add host group names to an array.
	for _, group := range jsonResults.Groups {
		groups = append(groups, group.Groupname)
	}

	if groups == nil {
		fmt.Println("Error, no host groups found in hosts.json")
		return
	}

	// User wants to add a host
	if *addHost {
		AddHost(jsonResults, groups, configFile)
		return
	}

	// User wants to delete a host
	if *delHost {
		DelHost(jsonResults, groups, configFile)
		return
	}

	// Adding the all option to Groups
	groups = append(groups, "All")

	// Bring up the menu for the user to select the group to filter to.
	var selectGroup string
	groupPrompt := &survey.Select{
		Message: "Select a Device Group:",
		Options: groups,
	}
	survey.AskOne(groupPrompt, &selectGroup, survey.WithPageSize(10))

	// Iterate through the JSON data and store a list of devices, hostnames and users to arrays

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
		// Otherwise, just add the hosts for the group selected
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
