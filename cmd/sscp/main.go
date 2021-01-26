package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"os/exec"
	"os/signal"

	"github.com/mitchellh/go-homedir"
	"github.com/ysugimoto/sssh"
)

var configFile string
var upload bool

func main() {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln(err)
	}
	flag.StringVar(&configFile, "f", home+"/.ssh/config", "Load ssh configration file path")
	flag.BoolVar(&upload, "u", false, "Inverse in-out, it means upload file")
	flag.Parse()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		os.Exit(1)
	}()

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
		selections[i] = fmt.Sprintf(
			"%-"+fmt.Sprint(max)+"s [scp %s]",
			h.Name,
			command(h),
		)
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

	var remote, local string
	var cmd *exec.Cmd
	if !upload {
		fmt.Println("==== Download Mode (remote -> local) ====")
		remote = getRemotePath(host)
		local = getLocalPath()
		cmd = exec.Command("scp", host.Name+":"+remote, local)
	} else {
		fmt.Println("==== Upload Mode (local -> remote) ====")
		local = getLocalPath()
		remote = getRemotePath(host)
		cmd = exec.Command("scp", local, host.Name+":"+remote)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func command(h *sssh.Host) string {
	var c []string
	if h.IdentityFile != "" {
		c = append(c, "-i "+h.IdentityFile)
	}
	if h.Port != "" {
		c = append(c, "-P "+h.Port)
	}
	if h.User != "" {
		c = append(c, h.User+"@"+h.HostName)
	} else {
		c = append(c, h.HostName)
	}
	return strings.Join(c, " ")
}

func getRemotePath(host *sssh.Host) string {
	var remote string
	for {
		fmt.Printf("Remote [%s@%s]: ", host.User, host.HostName)
		fmt.Scanln(&remote)
		remote = strings.TrimSpace(remote)
		if remote == "" {
			continue
		}
		break
	}
	return strings.ReplaceAll(remote, " ", "\\ ")
}

func getLocalPath() string {
	var local string
	wd, _ := os.Getwd()
	for {
		fmt.Printf("Local [%s]: ", wd)
		fmt.Scanln(&local)
		local = strings.TrimSpace(local)
		if local == "" {
			continue
		}
		break
	}
	return strings.ReplaceAll(local, " ", "\\ ")
}
