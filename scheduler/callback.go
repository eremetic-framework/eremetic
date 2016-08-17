package scheduler

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"

	"github.com/klarna/eremetic/types"
)

type callbackData struct {
	Time   int64  `json:"time"`
	Status string `json:"status"`
	TaskID string `json:"task_id"`
}

// NotifyCallback handles posting a JSON back to the URI given with the task.
func NotifyCallback(task *types.EremeticTask) {
	if len(task.CallbackURI) == 0 {
		return
	}

	cbData := callbackData{
		Time:   task.Status[len(task.Status)-1].Time,
		Status: string(task.Status[len(task.Status)-1].Status),
		TaskID: task.ID,
	}

	body, err := json.Marshal(cbData)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"task_id":      task.ID,
			"callback_uri": task.CallbackURI,
		}).Error("Unable to create callback message")
		return
	}

	go func() {
		_, err = http.Post(task.CallbackURI, "application/json", bytes.NewBuffer(body))

		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"task_id":      task.ID,
				"callback_uri": task.CallbackURI,
			}).Error("Unable to POST to Callback URI")
		} else {
			logrus.WithFields(logrus.Fields{
				"task_id":      task.ID,
				"callback_uri": task.CallbackURI,
			}).Debug("Sent callback")
		}
	}()
}
