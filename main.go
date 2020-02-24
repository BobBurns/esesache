package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type Options struct {
	key   string
	value string
}
type Host struct {
	name     string
	hostname string
	user     string
	options  []Options
}

func main() {
	file, err := os.Open("/home/bob/.ssh/config")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hosts []Host

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		if !unicode.IsSpace(rune(line[0])) {
			fields := strings.Fields(line)
			host := Host{}
			if fields[0] == "Host" {
				host.name = fields[1]
			} else {
				log.Println("Must have first field as Host")
				continue
			}
			for scanner.Scan() {
				line = scanner.Text()
				fields = strings.Fields(line)
				if len(fields) == 0 {
					break
				}
				if fields[0] == "User" {
					host.user = fields[1]
				} else if fields[0] == "Hostname" {
					host.hostname = fields[1]
				} else {
					opt := Options{
						key:   fields[0],
						value: fields[1],
					}
					//	fmt.Println("debug", opt)
					host.options = append(host.options, opt)
				}

			}
			hosts = append(hosts, host)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// get selection
	var choices []string
	for _, h := range hosts {
		choices = append(choices, h.name)
	}

	result, input := GetSelection(choices)
	if result != "ok" {

		fmt.Println(result)
		os.Exit(0)
	}

	hostArg := fmt.Sprintf("%s@%s", hosts[input].user, hosts[input].hostname)
	var sshArgs []string
	if len(hosts[input].options) > 0 {
		for _, o := range hosts[input].options {
			s := "-o"
			sshArgs = append(sshArgs, s)
			s = fmt.Sprintf("%s=%s ", o.key, o.value)
			sshArgs = append(sshArgs, s)

		}
	}
	sshArgs = append(sshArgs, hostArg)

	fmt.Printf("ssh ")
	for _, a := range sshArgs {
		fmt.Printf("%s ", a)
	}
	fmt.Println()
	//	os.Exit(0)
	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
