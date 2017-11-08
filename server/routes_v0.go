package server

import (
	"github.com/cybricio/eremetic/api"
	"github.com/cybricio/eremetic/config"
)

func apiV0Routes(h Handler, conf *config.Config) Routes {
	return Routes{
		Route{
			Name:    "AddTask",
			Method:  "POST",
			Pattern: "/task",
			Handler: h.AddTask(conf, api.V0),
		},
		Route{
			Name:    "Status",
			Method:  "GET",
			Pattern: "/task/{taskId}",
			Handler: h.GetTaskInfo(conf, api.V0),
		},
		Route{
			Name:    "STDOUT",
			Method:  "GET",
			Pattern: "/task/{taskId}/stdout",
			Handler: h.GetFromSandbox("stdout", api.V0),
		},
		Route{
			Name:    "STDERR",
			Method:  "GET",
			Pattern: "/task/{taskId}/stderr",
			Handler: h.GetFromSandbox("stderr", api.V0),
		},
		Route{
			Name:    "Kill",
			Method:  "POST",
			Pattern: "/task/{taskId}/kill",
			Handler: h.KillTask(conf, api.V0),
		},
		Route{
			Name:    "Delete",
			Method:  "DELETE",
			Pattern: "/task/{taskId}",
			Handler: h.DeleteTask(conf, api.V0),
		},
		Route{
			Name:    "ListRunningTasks",
			Method:  "GET",
			Pattern: "/task",
			Handler: h.ListRunningTasks(api.V0),
		},
		Route{
			Name:    "Version",
			Method:  "GET",
			Pattern: "/version",
			Handler: h.Version(conf, api.V0),
		},
	}
}
