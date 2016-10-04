package mesos

import (
	"fmt"

	"github.com/gogo/protobuf/proto"

	"github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"

	"github.com/klarna/eremetic"
)

func createTaskInfo(task eremetic.Task, offer *mesosproto.Offer) (eremetic.Task, *mesosproto.TaskInfo) {
	task.FrameworkID = *offer.FrameworkId.Value
	task.SlaveID = *offer.SlaveId.Value
	task.Hostname = *offer.Hostname
	task.AgentIP = offer.GetUrl().GetAddress().GetIp()
	task.AgentPort = offer.GetUrl().GetAddress().GetPort()

	portMapping, portResources := buildPorts(task, offer)
	env := buildEnvironment(task, portMapping)

	taskInfo := &mesosproto.TaskInfo{
		TaskId:  &mesosproto.TaskID{Value: proto.String(task.ID)},
		SlaveId: offer.SlaveId,
		Name:    proto.String(task.Name),
		Command: buildCommandInfo(task, env),
		Container: &mesosproto.ContainerInfo{
			Type: mesosproto.ContainerInfo_DOCKER.Enum(),
			Docker: &mesosproto.ContainerInfo_DockerInfo{
				Image:          proto.String(task.Image),
				ForcePullImage: proto.Bool(task.ForcePullImage),
				PortMappings:   portMapping,
				Network:        mesosproto.ContainerInfo_DockerInfo_BRIDGE.Enum(),
			},
			Volumes: buildVolumes(task),
		},
		Resources: []*mesosproto.Resource{
			mesosutil.NewScalarResource("cpus", task.TaskCPUs),
			mesosutil.NewScalarResource("mem", task.TaskMem),
			mesosutil.NewRangesResource("ports", portResources),
		},
	}
	return task, taskInfo
}

func buildEnvironment(task eremetic.Task, portMappings []*mesosproto.ContainerInfo_DockerInfo_PortMapping) *mesosproto.Environment {
	var environment []*mesosproto.Environment_Variable
	for k, v := range task.Environment {
		environment = append(environment, &mesosproto.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}
	for k, v := range task.MaskedEnvironment {
		environment = append(environment, &mesosproto.Environment_Variable{
			Name:  proto.String(k),
			Value: proto.String(v),
		})
	}
	for i, m := range portMappings {
		environment = append(environment, &mesosproto.Environment_Variable{
			Name:  proto.String(fmt.Sprintf("PORT%d", i)),
			Value: proto.String(fmt.Sprintf("%d", *m.HostPort)),
		})
	}
	if len(portMappings) > 0 {
		environment = append(environment, &mesosproto.Environment_Variable{
			Name:  proto.String("PORT"),
			Value: proto.String(fmt.Sprintf("%d", *portMappings[0].HostPort)),
		})
	}

	environment = append(environment, &mesosproto.Environment_Variable{
		Name:  proto.String("MESOS_TASK_ID"),
		Value: proto.String(task.ID),
	})

	return &mesosproto.Environment{
		Variables: environment,
	}
}

func buildVolumes(task eremetic.Task) []*mesosproto.Volume {
	var volumes []*mesosproto.Volume
	for _, v := range task.Volumes {
		volumes = append(volumes, &mesosproto.Volume{
			Mode:          mesosproto.Volume_RW.Enum(),
			ContainerPath: proto.String(v.ContainerPath),
			HostPath:      proto.String(v.HostPath),
		})
	}

	return volumes
}

func buildPorts(task eremetic.Task, offer *mesosproto.Offer) ([]*mesosproto.ContainerInfo_DockerInfo_PortMapping, []*mesosproto.Value_Range) {
	var resources []*mesosproto.Value_Range
	var mappings []*mesosproto.ContainerInfo_DockerInfo_PortMapping

	if len(task.Ports) == 0 {
		return mappings, resources
	}

	leftToAssign := len(task.Ports)

	for _, rsrc := range offer.Resources {
		if *rsrc.Name != "ports" {
			continue
		}

		for _, rng := range rsrc.Ranges.Range {
			if leftToAssign == 0 {
				break
			}

			start, end := *rng.Begin, *rng.Begin

			for hport := int(*rng.Begin); hport <= int(*rng.End); hport++ {
				if leftToAssign == 0 {
					break
				}

				leftToAssign--

				tport := &task.Ports[leftToAssign]
				tport.HostPort = uint32(hport)

				if tport.ContainerPort == 0 {
					tport.ContainerPort = tport.HostPort
				}

				end = uint64(hport + 1)

				mappings = append(mappings, &mesosproto.ContainerInfo_DockerInfo_PortMapping{
					ContainerPort: proto.Uint32(tport.ContainerPort),
					HostPort:      proto.Uint32(tport.HostPort),
					Protocol:      proto.String(tport.Protocol),
				})
			}

			if start != end {
				resources = append(resources, mesosutil.NewValueRange(start, end))
			}
		}
	}

	return mappings, resources
}

func buildURIs(task eremetic.Task) []*mesosproto.CommandInfo_URI {
	var uris []*mesosproto.CommandInfo_URI
	for _, v := range task.FetchURIs {
		uris = append(uris, &mesosproto.CommandInfo_URI{
			Value:      proto.String(v.URI),
			Extract:    proto.Bool(v.Extract),
			Executable: proto.Bool(v.Executable),
			Cache:      proto.Bool(v.Cache),
		})
	}

	return uris
}

func buildCommandInfo(task eremetic.Task, env *mesosproto.Environment) *mesosproto.CommandInfo {
	commandInfo := &mesosproto.CommandInfo{
		User:        proto.String(task.User),
		Environment: env,
		Uris:        buildURIs(task),
	}

	if task.Command != "" {
		commandInfo.Shell = proto.Bool(true)
		commandInfo.Value = &task.Command
	} else {
		commandInfo.Shell = proto.Bool(false)
		commandInfo.Arguments = task.Args
	}

	return commandInfo
}
