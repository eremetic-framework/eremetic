package zook

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/dmuth/google-go-log4go"

	"github.com/samuel/go-zookeeper/zk"
)

// DiscoverMaster attempts to find the Mesos Master from a Zookeeper URI
func DiscoverMaster(zkString string) string {
	uris, path := splitURI(zkString)
	z, _, err := zk.Connect(uris, time.Second)
	handleError(err)

	c, _, err := z.Children(path)
	handleError(err)

	var node string
	for _, n := range c {
		if !strings.HasPrefix(n, "json.info_") {
			continue
		}

		if node == "" || strings.Compare(n, node) < 0 {
			node = n
		}
	}

	if node == "" {
		log.Error("Could not discover master")
		os.Exit(1)
	}

	var resp zkChild
	data, _, err := z.Get(path + "/" + node)
	handleError(err)
	err = json.Unmarshal(data, &resp)
	handleError(err)

	log.Debugf(
		"Found node with IP: %s, Hostname: %s, Port: %d",
		resp.Address.IP, resp.Address.Hostname, resp.Address.Port)

	return fmt.Sprintf("%s:%d", resp.Address.IP, resp.Address.Port)

}

func splitURI(uri string) ([]string, string) {
	uri = strings.Replace(uri, "zk://", "", 2)
	split := strings.Split(uri, "/")
	arr := strings.Split(split[0], ",")
	path := "/" + split[1]

	return arr, path
}

type zkChild struct {
	Address zkAddress `json:"address"`
}

type zkAddress struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
}

func handleError(err error) {
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}
