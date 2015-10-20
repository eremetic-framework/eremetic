package zook

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
)

// DiscoverMaster attempts to find the Mesos Master from a Zookeeper URI
func DiscoverMaster(zkString string) string {
	uris, path := splitURI(zkString)
	z, _, err := zk.Connect(uris, time.Second)
	handleError(err)

	c, _, err := z.Children(path)
	handleError(err)
	fmt.Println(c)

	var nodes []zkChild

	for _, child := range c {
		var resp zkChild
		data, _, _ := z.Get(path + "/" + child)
		err = json.Unmarshal(data, &resp)
		if err != nil {
			continue
		}
		log.Printf("Found node with IP: %s, Hostname: %s, Port: %d", resp.Address.IP, resp.Address.Hostname, resp.Address.Port)
		nodes = append(nodes, resp)
	}
	return fmt.Sprintf("%s:%d", nodes[0].Address.IP, nodes[0].Address.Port)
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
		log.Fatal(err)
	}
}
