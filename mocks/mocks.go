package mocks

import (
	"errors"

	"github.com/klarna/eremetic/types"
)

type Scheduler struct {
	NextError *error
}

func (s *Scheduler) ScheduleTask(request types.Request) (string, error) {
	if err := s.NextError; err != nil {
		s.NextError = nil
		return "", *err
	}
	return "eremetic-task.mock", nil
}

type ErrorReader struct{}

func (r *ErrorReader) Read(p []byte) (int, error) {
	return 0, errors.New("oh no")
}
