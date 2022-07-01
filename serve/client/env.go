package client

import (
	"fmt"
	"io/ioutil"
	"strings"
)

var (
	medeaServer = "http://127.0.0.1:8630"
	medeaHost   = "127.0.0.1"
)

var (
	envTargetFile      = ".env"
	envMedeaServername = "MEDEA_SERVER"
	envMedeaHostname   = "MEDEA_HOST"
)

func env_update(val map[string]string) error {
	server := val["server"]
	host := val["host"]

	envStr := ""
	if len(server) > 0 {
		envStr = envStr + fmt.Sprintf("export %s=%s", envMedeaServername, server) + "\n"
	}

	if len(host) > 0 {
		envStr = envStr + fmt.Sprintf("export %s=%s", envMedeaHostname, host) + "\n"
	}

	err := ioutil.WriteFile(envTargetFile, []byte(envStr), 0644)
	if err != nil {
		panic(err)
	}

	return nil
}

func isFileExsit(path string) (bool, error) {
	finfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	if !finfo.IsDir() {
		return true, nil
	}

	return false, nil
}

func env_refresh() (map[string]string, error) {
	ok, err := isFileExsit(envTargetFile)
	if err != nil || !ok {
		m := make(map[string]string, 2)
		m[envMedeaServername] = medeaServer
		m[envMedeaHostname] = medeaHost
		return m, nil
	}

	content, err := ioutil.ReadFile(envTargetFile)
	if err != nil {
		panic(err)
	}

	s := strings.Split(string(content), "\n")
	m := make(map[string]string, 2)
	for _, v := range s {
		if len(v) == 0 {
			continue
		}
		val := strings.Split(v, " ")
		if len(val) != 2 {
			continue
		}
		item := strings.Split(val[1], "=")
		if len(item) != 2 {
			continue
		}

		m[item[0]] = item[1]
	}

	return m, nil
}

func globalEnvironmentUpdate() {
	m, err := env_refresh()
	if err != nil {
		return
	}

	if len(m[envMedeaServername]) != 0 {
		medeaServer = m[envMedeaServername]
	}

	if len(m[envMedeaHostname]) != 0 {
		medeaHost = m[envMedeaHostname]
	}
}
