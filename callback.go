package eremetic

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// CallbackData holds information about the status update.
type CallbackData struct {
	Time   int64  `json:"time"`
	Status string `json:"status"`
	TaskID string `json:"task_id"`
}

// NotifyCallback handles posting a JSON back to the URI given with the task.
func NotifyCallback(task *Task) {
	if len(task.CallbackURI) == 0 {
		return
	}

	if len(task.Status) == 0 {
		return
	}

	status := task.Status[len(task.Status)-1]

	data := CallbackData{
		Time:   status.Time,
		Status: status.Status.String(),
		TaskID: task.ID,
	}

	body, err := json.Marshal(data)
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

			return
		}

		logrus.WithFields(logrus.Fields{
			"task_id":      task.ID,
			"callback_uri": task.CallbackURI,
		}).Debug("Sent callback")
	}()
}
