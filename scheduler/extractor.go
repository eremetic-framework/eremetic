package scheduler

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type mounts struct {
	Mounts []dockerMounts `json:"Mounts"`
}

type dockerMounts struct {
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
}

func extractSandboxPath(status *mesos.TaskStatus) (string, error) {
	var mounts []mounts

	if len(status.Data) == 0 {
		logrus.Debug("No Data in task status.")
		return "", nil
	}

	err := json.Unmarshal(status.Data, &mounts)

	if err != nil {
		logrus.WithError(err).Error("Task status data contained invalid JSON.")
		return "", err
	}

	for _, m := range mounts {
		for _, dm := range m.Mounts {
			if dm.Destination == "/mnt/mesos/sandbox" {
				return dm.Source, nil
			}
		}
	}

	logrus.Debug("No sandbox mount found in task status data.")
	return "", nil
}
