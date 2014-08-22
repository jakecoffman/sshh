package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"code.google.com/p/go.crypto/ssh"
	"github.com/howeyc/gopass"
)

type Host struct {
	hostname string
	client   *ssh.Client
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	username, hostnames, pass := parse()

	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(string(pass)),
		},
	}

	hosts := []*Host{}
	for _, host := range hostnames {
		client, err := ssh.Dial("tcp", host+":22", config)
		if err != nil {
			log.Fatal("Failed to dial: ", err)
		}
		hosts = append(hosts, &Host{host, client})
	}

	lock := make(chan string)

	for {
		fmt.Print("$ ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal("Error getting user input: ", err)
		}

		for _, host := range hosts {
			go func(host *Host) {
				// Each ClientConn can support multiple interactive sessions,
				// represented by a Session.
				session, err := host.client.NewSession()
				if err != nil {
					log.Fatal("Failed to create session: ", err)
				}

				// Once a Session is created, you can execute a single command on
				// the remote side using the Run method.
				var b bytes.Buffer
				session.Stdout = &b
				if err := session.Run(input); err != nil {
					lock <- fmt.Sprintf("%v\n%v", host.hostname, err)
				} else {
					lock <- fmt.Sprintf("%v\n%v", host.hostname, b.String())
				}
			}(host)
		}

		for _ = range hosts {
			fmt.Println(<-lock)
		}
	}
}

func parse() (username string, hosts []string, pass string) {
	username = os.Args[1]
	hosts = strings.Split(os.Args[2], ",")

	fmt.Printf("Password: ")
	pass = string(gopass.GetPasswd())
	return
}

func usage() {
	fmt.Println("Usage:\n\tsshh [username] [comma-separated-hosts]")
}
