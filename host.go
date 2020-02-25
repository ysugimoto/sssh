package sssh

import (
	"fmt"
	"strings"
)

type Host struct {
	Name         string
	HostName     string
	User         string
	Port         string
	IdentityFile string
}

func (h *Host) Format(cmd string, align int) string {
	return fmt.Sprintf(
		"%-"+fmt.Sprint(align)+"s [%s]",
		h.Name,
		h.Command(cmd),
	)
}

func (h *Host) Command(cmd string) string {
	var c []string
	if h.IdentityFile != "" {
		c = append(c, "-i "+h.IdentityFile)
	}
	if h.Port != "" {
		c = append(c, "-p "+h.Port)
	}
	if h.User != "" {
		c = append(c, h.User+"@"+h.HostName)
	} else {
		c = append(c, h.HostName)
	}
	return strings.Join(c, " ")
}

type Hosts []*Host

func (hs Hosts) MaxSize() int {
	var max int
	for _, v := range hs {
		if len(v.Name) > max {
			max = len(v.Name)
		}
	}
	return max
}

func (hs Hosts) Find(name string) *Host {
	for _, h := range hs {
		if h.Name == name {
			return h
		}
	}
	return nil
}
