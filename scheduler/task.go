package scheduler

import (
	"github.com/golang/protobuf/proto"
	"github.com/klarna/eremetic/types"
	mesos "github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

func createTaskInfo(task types.EremeticTask, offer *mesos.Offer) (types.EremeticTask, *mesos.TaskInfo) {
	task.FrameworkId = *offer.FrameworkId.Value
	task.SlaveId = *offer.SlaveId.Value
	task.Hostname = *offer.Hostname
	task.AgentIP = offer.GetUrl().GetAddress().GetIp()
	task.AgentPort = offer.GetUrl().GetAddress().GetPort()

	taskInfo := &mesos.TaskInfo{
		TaskId:  &mesos.TaskID{Value: proto.String(task.ID)},
		SlaveId: offer.SlaveId,
		Name:    proto.String(task.Name),
		Command: buildCommandInfo(task),
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image:          proto.String(task.Image),
				ForcePullImage: proto.Bool(task.ForcePullImage),
			},
			Volumes: buildVolumes(task),
		},
		Resources: []*mesos.Resource{
			mesosutil.NewScalarResource("cpus", task.TaskCPUs),
			mesosutil.NewScalarResource("mem", task.TaskMem),
		},
	}
	return task, taskInfo
}

func buildEnvironment(task types.EremeticTask) []*mesos.Environment_Variable {
	var environment []*mesos.Environment_Variable
	for k, v := range task.Environment {
		environment = append(environment, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}
	for k, v := range task.MaskedEnvironment {
		environment = append(environment, &mesos.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}

	environment = append(environment, &mesos.Environment_Variable{
		Name:  proto.String("MESOS_TASK_ID"),
		Value: proto.String(task.ID),
	})

	return environment
}

func buildVolumes(task types.EremeticTask) []*mesos.Volume {
	var volumes []*mesos.Volume
	for _, v := range task.Volumes {
		volumes = append(volumes, &mesos.Volume{
			Mode:          mesos.Volume_RW.Enum(),
			ContainerPath: proto.String(v.ContainerPath),
			HostPath:      proto.String(v.HostPath),
		})
	}

	return volumes
}

func buildURIs(task types.EremeticTask) []*mesos.CommandInfo_URI {
	var uris []*mesos.CommandInfo_URI
	for _, v := range task.FetchURIs {
		uris = append(uris, &mesos.CommandInfo_URI{
			Value:      proto.String(v.URI),
			Extract:    proto.Bool(v.Extract),
			Executable: proto.Bool(v.Executable),
			Cache:      proto.Bool(v.Cache),
		})
	}

	return uris
}

func buildCommandInfo(task types.EremeticTask) *mesos.CommandInfo {
	commandInfo := &mesos.CommandInfo{
		User: proto.String(task.User),
		Environment: &mesos.Environment{
			Variables: buildEnvironment(task),
		},
		Uris: buildURIs(task),
	}

	if task.Command != "" {
		commandInfo.Value = &task.Command
	} else {
		commandInfo.Shell = proto.Bool(false)
	}

	return commandInfo
}
