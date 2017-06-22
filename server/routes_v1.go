package server

import (
	"github.com/eremetic-framework/eremetic/api"
	"github.com/eremetic-framework/eremetic/config"
)

func apiV1Routes(h Handler, conf *config.Config) Routes {
	return Routes{
		Route{
			Name:    "AddTask",
			Method:  "POST",
			Pattern: "/api/v1/task",
			Handler: h.AddTask(conf, api.V1),
		},
		Route{
			Name:    "Status",
			Method:  "GET",
			Pattern: "/api/v1/task/{taskId}",
			Handler: h.GetTaskInfo(conf, api.V1),
		},
		Route{
			Name:    "STDOUT",
			Method:  "GET",
			Pattern: "/api/v1/task/{taskId}/stdout",
			Handler: h.GetFromSandbox("stdout", api.V1),
		},
		Route{
			Name:    "STDERR",
			Method:  "GET",
			Pattern: "/api/v1/task/{taskId}/stderr",
			Handler: h.GetFromSandbox("stderr", api.V1),
		},
		Route{
			Name:    "Kill",
			Method:  "POST",
			Pattern: "/api/v1/task/{taskId}/kill",
			Handler: h.KillTask(conf, api.V1),
		},
		Route{
			Name:    "Delete",
			Method:  "DELETE",
			Pattern: "/api/v1/task/{taskId}",
			Handler: h.DeleteTask(conf, api.V1),
		},
		Route{
			Name:    "ListRunningTasks",
			Method:  "GET",
			Pattern: "/api/v1/task",
			Handler: h.ListRunningTasks(api.V1),
		},
		Route{
			Name:    "Version",
			Method:  "GET",
			Pattern: "/api/v1/version",
			Handler: h.Version(conf, api.V1),
		},
	}
}
