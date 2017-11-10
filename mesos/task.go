package mesos

import (
	"fmt"

	"github.com/mesos/mesos-go/api/v1/lib"

	"github.com/eremetic-framework/eremetic"
)

func createTaskInfo(task eremetic.Task, offer *mesos.Offer) (eremetic.Task, *mesos.TaskInfo) {
	task.FrameworkID = offer.FrameworkID.GetValue()
	task.AgentID = offer.AgentID.GetValue()
	task.Hostname = offer.Hostname
	address := offer.GetURL().GetAddress()
	task.AgentIP = address.GetIP()
	task.AgentPort = address.GetPort()

	network := buildNetwork(task)
	dockerCliParameters := buildDockerCliParameters(task)
	portMapping, portResources := buildPorts(task, network, offer)
	env := buildEnvironment(task, portMapping)

	resources := []mesos.Resource{
		*mesos.BuildResource().Name("cpus").Scalar(task.TaskCPUs).Resource,
		*mesos.BuildResource().Name("mem").Scalar(task.TaskMem).Resource,
	}
	if len(portResources) > 0 {
		resources = append(resources, *mesos.BuildResource().Name("ports").Ranges(portResources).Resource)
	}

	taskInfo := &mesos.TaskInfo{
		TaskID:  mesos.TaskID{Value: task.ID},
		AgentID: offer.AgentID,
		Name:    task.Name,
		Command: buildCommandInfo(task, env),
		Container: &mesos.ContainerInfo{
			Type: mesos.ContainerInfo_DOCKER.Enum(),
			Docker: &mesos.ContainerInfo_DockerInfo{
				Image:          task.Image,
				ForcePullImage: &task.ForcePullImage,
				Privileged:     &task.Privileged,
				Network:        network,
				PortMappings:   portMapping,
				Parameters:     dockerCliParameters,
			},
			Volumes: buildVolumes(task),
		},
		Resources: resources,
	}
	return task, taskInfo
}

func buildDockerCliParameters(task eremetic.Task) []mesos.Parameter {
	//To be able to move away from docker CLI in future, parameters aren't fully exposed to the API
	params := make(map[string]string)
	if task.DNS != "" {
		params["dns"] = task.DNS
	}
	var parameters []mesos.Parameter
	for k, v := range params {
		parameters = append(parameters, mesos.Parameter{
			Key:   k,
			Value: v,
		})
	}
	if len(task.VolumesFrom) > 0 {
		for _, containerName := range task.VolumesFrom {
			if containerName == "" {
				continue
			}
			parameters = append(parameters, mesos.Parameter{
				Key:   "volumes-from",
				Value: containerName,
			})
		}
	}
	return parameters
}

func buildNetwork(task eremetic.Task) *mesos.ContainerInfo_DockerInfo_Network {
	if task.Network == "" {
		return mesos.ContainerInfo_DockerInfo_BRIDGE.Enum()
	}
	return mesos.ContainerInfo_DockerInfo_Network(mesos.ContainerInfo_DockerInfo_Network_value[task.Network]).Enum()
}

func buildEnvironment(task eremetic.Task, portMappings []mesos.ContainerInfo_DockerInfo_PortMapping) *mesos.Environment {
	var environment []mesos.Environment_Variable
	for k, v := range task.Environment {
		environment = append(environment, mesos.Environment_Variable{
			Name:  k,
			Value: v,
		})
	}
	for k, v := range task.MaskedEnvironment {
		environment = append(environment, mesos.Environment_Variable{
			Name:  k,
			Value: v,
		})
	}
	for i, m := range portMappings {
		environment = append(environment, mesos.Environment_Variable{
			Name:  fmt.Sprintf("PORT%d", i),
			Value: fmt.Sprintf("%d", m.HostPort),
		})
	}
	if len(portMappings) > 0 {
		environment = append(environment, mesos.Environment_Variable{
			Name:  "PORT",
			Value: fmt.Sprintf("%d", portMappings[0].HostPort),
		})
	}

	environment = append(environment, mesos.Environment_Variable{
		Name:  "MESOS_TASK_ID",
		Value: task.ID,
	})

	return &mesos.Environment{
		Variables: environment,
	}
}

func buildVolumes(task eremetic.Task) []mesos.Volume {
	var volumes []mesos.Volume
	for _, v := range task.Volumes {
		volumes = append(volumes, mesos.Volume{
			Mode:          mesos.RW.Enum(),
			ContainerPath: v.ContainerPath,
			HostPath:      &v.HostPath,
		})
	}

	return volumes
}

func buildPorts(task eremetic.Task, network *mesos.ContainerInfo_DockerInfo_Network, offer *mesos.Offer) ([]mesos.ContainerInfo_DockerInfo_PortMapping, mesos.Ranges) {
	var mappings []mesos.ContainerInfo_DockerInfo_PortMapping
	resources := mesos.BuildRanges()

	if len(task.Ports) == 0 || *network == mesos.ContainerInfo_DockerInfo_HOST {
		return mappings, resources.Ranges
	}

	leftToAssign := len(task.Ports)

	for _, rsrc := range offer.Resources {
		if rsrc.Name != "ports" {
			continue
		}

		for _, rng := range rsrc.Ranges.Range {
			if leftToAssign == 0 {
				break
			}

			start, end := rng.Begin, rng.Begin

			for hport := int(rng.Begin); hport <= int(rng.End); hport++ {
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

				mappings = append(mappings, mesos.ContainerInfo_DockerInfo_PortMapping{
					ContainerPort: tport.ContainerPort,
					HostPort:      tport.HostPort,
					Protocol:      &tport.Protocol,
				})
			}

			if start != end {
				resources = resources.Span(start, end)
			}
		}
	}

	return mappings, resources.Ranges
}

func buildURIs(task eremetic.Task) []mesos.CommandInfo_URI {
	var uris []mesos.CommandInfo_URI
	for _, v := range task.FetchURIs {
		uris = append(uris, mesos.CommandInfo_URI{
			Value:      v.URI,
			Extract:    &v.Extract,
			Executable: &v.Executable,
			Cache:      &v.Cache,
		})
	}

	return uris
}

func buildCommandInfo(task eremetic.Task, env *mesos.Environment) *mesos.CommandInfo {
	commandInfo := &mesos.CommandInfo{
		User:        &task.User,
		Environment: env,
		URIs:        buildURIs(task),
	}

	if task.Command != "" {
		shell := true
		commandInfo.Shell = &shell
		commandInfo.Value = &task.Command
	} else {
		shell := false
		commandInfo.Shell = &shell
		commandInfo.Arguments = task.Args
	}

	return commandInfo
}
