package sssh

import (
	"bufio"
	"errors"
	"io"
	"strings"

	"github.com/ysugimoto/cho"
)

func Load(r io.Reader) (Hosts, error) {
	var hosts Hosts
	var h *Host

	reader := bufio.NewReader(r)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return hosts, nil
			}
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "Host ") {
			if h != nil {
				hosts = append(hosts, h)
			}
			h = &Host{
				Name: strings.TrimPrefix(line, "Host "),
			}
			continue
		}
		kv := strings.SplitN(line, " ", 2)
		if len(kv) != 2 {
			return nil, errors.New("failed to parse configuration, invalid definition: " + line)
		} else if h == nil {
			return nil, errors.New("failed to parse configuration, unexpected definition: " + line)
		}
		switch strings.ToLower(kv[0]) {
		case "hostname":
			h.HostName = parseValue(kv[1])
		case "user":
			h.User = parseValue(kv[1])
		case "port":
			h.Port = parseValue(kv[1])
		case "identityfile":
			h.IdentityFile = parseValue(kv[1])
		}
	}
}

func parseValue(v string) string {
	v = strings.TrimSpace(v)
	if idx := strings.Index(v, "#"); idx > -1 {
		return v[0:idx]
	}
	return v
}

func Select(list []string) (result string) {
	ret := make(chan string, 1)
	terminate := make(chan struct{})
	go cho.Run(list, ret, terminate)

	for {
		select {
		case result = <-ret:
			return
		case <-terminate:
			return
		}
	}
}
