package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"

	"io/ioutil"
	"os/exec"

	"github.com/mitchellh/go-homedir"
	"github.com/ysugimoto/cho"
)

var sectionSplit = []byte("Host ")
var linefeed = []byte("\n")
var commentSignature = []byte("#")

type Host struct {
	Host         string
	Hostname     string
	User         string
	Port         string
	IdentityFile string
}

func (h Host) Format(align int) string {
	format := "%-" + fmt.Sprint(align) + "s [ssh"
	if h.IdentityFile != "" {
		format += " -i " + h.IdentityFile + " "
	}
	if h.Port != "" {
		format += " -p " + h.Port + " "
	}
	if h.User != "" {
		format += h.User + "@"
	}
	format += h.Hostname + "]"
	return fmt.Sprintf(format, h.Host)
}

var configFile string

func init() {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	flag.StringVar(&configFile, "f", home+"/.ssh/config", "Load ssh configration file path")
	flag.Parse()
}

func main() {
	if _, err := os.Stat(configFile); err != nil {
		panic(err)
	}
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(err)
	}

	sections := bytes.Split(config, sectionSplit)
	hosts := []Host{}
	max := 0
	for _, v := range sections {
		v = bytes.TrimSpace(v)
		if len(v) == 0 {
			continue
		}
		host := parseHost(v)
		hosts = append(hosts, host)
		if len(host.Host) > max {
			max = len(host.Host)
		}
	}
	list := []string{}
	for _, v := range hosts {
		list = append(list, v.Format(max+1))
	}
	ret := make(chan string, 1)
	terminate := make(chan struct{})
	go cho.Run(list, ret, terminate)
	result := ""
LOOP:
	for {
		select {
		case result = <-ret:
			break LOOP
		case <-terminate:
			os.Exit(1)
		}
	}
	spl := strings.SplitN(result, "[", 2)
	cmd := exec.Command("ssh", strings.TrimSpace(spl[0]))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func parseValue(v []byte) string {
	v = bytes.TrimSpace(v)
	idx := bytes.Index(v, commentSignature)
	if idx > -1 {
		v = v[0:idx]
	}
	return string(v)
}

func parseHost(buffer []byte) Host {
	h := Host{}
	lines := bytes.Split(buffer, linefeed)
	h.Host = string(lines[0])
	for _, line := range lines[1:] {
		line = bytes.TrimSpace(line)
		if bytes.HasPrefix(line, commentSignature) {
			continue
		}
		spl := bytes.SplitN(line, []byte(" "), 2)
		switch string(spl[0]) {
		case "HostName":
			h.Hostname = parseValue(spl[1])
		case "User":
			h.User = parseValue(spl[1])
		case "Port":
			h.Port = parseValue(spl[1])
		case "IdentityFile":
			h.IdentityFile = parseValue(spl[1])
		}
	}
	return h
}
