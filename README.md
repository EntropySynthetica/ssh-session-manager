# ssh-session-manager

A simple CLI app that lets you select which host from a list of hosts that you want to SSH into.  I have a large number of hosts to manage via SSH and I wanted a simple CLI app to let me quickly select which host I want to SSH into.  


## Installation

Change to your tmp directory with `cd /tmp`

Download ssm with ```curl -OL https://github.com/EntropySynthetica/ssh-session-manager/releases/download/v1.1.1/ssm-v1.1.1-linux-amd64.tar.gz```

Unzip ssm with `tar -xvzf ssm-v1.1.1-linux-amd64.tar.gz`

Add execute perms with `chmod +x ./ssm`

Move ssm to your executable path with `mv ./ssm /usr/local/bin`

test that everything works with `ssm -h`

## Usage

To run type `ssm` from the command line.  The program will ask you what group you would like to show hosts for, or you can select all to show all hosts.  After that select the host you would like to connect to and hit enter.  You can navigate with the arrow keys to select the group or host.  You can also type in the name of a group or host at anytime to filter the list.  

### First time Run,
If you have never run ssm before you will need to create an inventory file at ```/home/<username>/.config/ssm/hosts.json```

The program can create a sample file for you with the -new flag.  

### Operational Flags

`-addgroup <groupname>` Add a new group to the inventory file.  Hosts need to be saved within a group.  The name specified will be added. 

`-delgroup <groupname>` Remove a group.  Any hosts within this group will be removed

`-addhost` Add a host.  You will be asked what group to place the host in, then the Name, Hostname, and Username of the host.  

- Name can be anything and will be what you can search for when running the program.
- Hostname can be an IP or FQDN to the host.  
- Username is the username to log into that host with.  

Note, ssm does not handle password storage. You either need to enter a password when logging into a host, or use pre shared ssh keys for passwordless login.  

`-delhost` You will be asked to select a group and then the host to remove.  

`-version` Show version.

## Todo

* Add a Windows Powershell Build
* Add a Mac Build
* Add a session logging option


### Known Bugs to be fixed,
* The if the ~/.config/ssm folder doesn't exist and you run the -new command the program will act like it works, but do nothing. 
