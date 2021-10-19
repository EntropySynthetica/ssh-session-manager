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

	// Open the hosts file
	jsonFile, err := os.Open("hosts.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	// Unmarshal the JSON from the file
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var host Hosts
	json.Unmarshal(byteValue, &host)

	// Create emtpy array and dump the IPs into it.
	devices := []string{}
	for i := 0; i < len(host.Hosts); i++ {
		devices = append(devices, host.Hosts[i].Name)
	}

	// Bring up the menu for the user to select the device to connect to.
	var selectDevice int
	prompt := &survey.Select{
		Message: "Select a device to connect to:",
		Options: devices,
	}
	survey.AskOne(prompt, &selectDevice, survey.WithPageSize(10))

	// Run SSH and pass it the hostname of the device to connect to.
	fmt.Println("Connecting to " + host.Hosts[selectDevice].Name + " - " + host.Hosts[selectDevice].Ip)
	sshArgs := "ssh " + host.Hosts[selectDevice].User + "@" + host.Hosts[selectDevice].Ip
	cmd := exec.Command("bash", "-c", sshArgs)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
