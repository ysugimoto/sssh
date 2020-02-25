package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"

	"os/exec"

	"github.com/mitchellh/go-homedir"
	"github.com/ysugimoto/sssh"
)

var configFile string

func main() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}
	flag.StringVar(&configFile, "f", home+"/.ssh/config", "Load ssh configration file path")
	flag.Parse()

	if err := _main(); err != nil {
		log.Fatalln(err)
	}
}

func _main() error {
	if _, err := os.Stat(configFile); err != nil {
		return err
	}
	fp, err := os.Open(configFile)
	if err != nil {
		return err
	}
	defer fp.Close()

	hosts, err := sssh.Load(fp)
	if err != nil {
		return err
	}

	selections := make([]string, len(hosts))
	max := hosts.MaxSize()
	for i, h := range hosts {
		selections[i] = h.Format("ssh", max+1)
	}
	selected := sssh.Select(selections)
	if selected == "" {
		os.Exit(1)
	}
	parsed := strings.SplitN(selected, " ", 2)
	host := hosts.Find(strings.TrimSpace(parsed[0]))
	if host == nil {
		return errors.New("host not found")
	}
	cmd := exec.Command("ssh", host.Name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
