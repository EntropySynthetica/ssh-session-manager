package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/AlecAivazis/survey/v2"
)

type Hosts struct {
	Hosts []Host `json:"hosts"`
}

type Host struct {
	Name  string `json:"name"`
	Ip    string `json:"ip"`
	User  string `json:"user"`
	Group string `json:"group"`
}

func main() {

	// Open the inventory file
	jsonFile, err := os.Open("hosts.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	// Unmarshal the JSON from the inventory file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var host Hosts
	json.Unmarshal(byteValue, &host)

	//// Debug for loop to make sure the JSON file is getting parsed.
	// for i := 0; i < len(host.Hosts); i++ {
	// 	fmt.Println("Hostname: " + host.Hosts[i].Name)
	// 	fmt.Println("IP: " + host.Hosts[i].Ip)
	// 	fmt.Println("User: " + host.Hosts[i].User)
	// 	fmt.Println("Group: " + host.Hosts[i].Group)
	// }

	// Create emtpy array and dump the IPs into it.
	devices := []string{}
	for i := 0; i < len(host.Hosts); i++ {
		devices = append(devices, host.Hosts[i].Ip)
	}

	// Bring up the menu for the user to select the device to connect to.
	ssh_hostname := ""
	prompt := &survey.Select{
		Message: "Select a device to connect to:",
		Options: devices,
	}
	survey.AskOne(prompt, &ssh_hostname)

	// Run SSH and pass it the hostname of the device to connect to.
	fmt.Println("Connecting to " + ssh_hostname)
	sshArgs := "ssh erica@" + ssh_hostname
	cmd := exec.Command("bash", "-c", sshArgs)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
