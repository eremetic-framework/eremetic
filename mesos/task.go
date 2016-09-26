package mesos

import (
	"github.com/gogo/protobuf/proto"
	"github.com/klarna/eremetic"
	"github.com/mesos/mesos-go/mesosproto"
	"github.com/mesos/mesos-go/mesosutil"
)

func createTaskInfo(task eremetic.Task, offer *mesosproto.Offer) (eremetic.Task, *mesosproto.TaskInfo) {
	task.FrameworkId = *offer.FrameworkId.Value
	task.SlaveId = *offer.SlaveId.Value
	task.Hostname = *offer.Hostname
	task.AgentIP = offer.GetUrl().GetAddress().GetIp()
	task.AgentPort = offer.GetUrl().GetAddress().GetPort()

	portMapping, portResources := buildPorts(task, offer)

	taskInfo := &mesosproto.TaskInfo{
		TaskId:  &mesosproto.TaskID{Value: proto.String(task.ID)},
		SlaveId: offer.SlaveId,
		Name:    proto.String(task.Name),
		Command: buildCommandInfo(task),
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

func buildEnvironment(task eremetic.Task) []*mesosproto.Environment_Variable {
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

	environment = append(environment, &mesosproto.Environment_Variable{
		Name:  proto.String("MESOS_TASK_ID"),
		Value: proto.String(task.ID),
	})

	return environment
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
	var portResources []*mesosproto.Value_Range
	var portMapping []*mesosproto.ContainerInfo_DockerInfo_PortMapping

	if len(task.Ports) > 0 {
		lastIndex := len(task.Ports)

		for _, v := range offer.Resources {
			if lastIndex == 0 {
				break
			}

			if *v.Name != "ports" {
				continue
			}

			for _, p_v := range v.Ranges.Range {
				if lastIndex == 0 {
					break
				}

				startPort, endPort := *p_v.Begin, int(*p_v.Begin)
				for portnumber := int(*p_v.Begin); portnumber <= int(*p_v.End); portnumber++ {
					if lastIndex == 0 {
						break
					}

					lastIndex--
					ask_port := &task.Ports[lastIndex]

					if ask_port.ContainerPort == 0 {
						continue
					}

					endPort = portnumber + 1

					ask_port.HostPort = uint32(portnumber)

					portMapping = append(portMapping, &mesosproto.ContainerInfo_DockerInfo_PortMapping{
						ContainerPort: proto.Uint32(ask_port.ContainerPort),
						HostPort:      proto.Uint32(ask_port.HostPort),
						Protocol:      proto.String(ask_port.Protocol),
					})

				}
				if int(startPort) != endPort {
					portResources = append(portResources, mesosutil.NewValueRange(startPort, uint64(endPort)))
				}
			}
		}
	}

	return portMapping, portResources
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

func buildCommandInfo(task eremetic.Task) *mesosproto.CommandInfo {
	commandInfo := &mesosproto.CommandInfo{
		User: proto.String(task.User),
		Environment: &mesosproto.Environment{
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
