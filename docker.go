package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"log"
	"net/http/httputil"
	"net/url"
	"strings"
)

type DockerClient struct {
	client *client.Client
}

type RunningContainer struct {
	// container name like "awesome-web-service"
	Name string
	// status of container is healthy
	healthy   bool
	container types.Container
	// whether established docker network connection between target and karja.
	connected bool
	proxy     *httputil.ReverseProxy
}

func (c *DockerClient) fetchContainers() (ret []RunningContainer, err error) {
	containers, err := c.client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ctr := range containers {
		fmt.Printf("%s %v (status: %s)\n", ctr.ID, ctr.Ports, ctr.Status)
		if len(ctr.Ports) > 0 && len(ctr.Names) > 0 && strings.HasPrefix(ctr.Names[0], "/") {
			// ctr.Names starts with "/"
			name := strings.TrimPrefix(ctr.Names[0], "/")
			healthy := ctr.State == "running"
			ret = append(ret, RunningContainer{name, healthy, ctr, false, nil})
		}
	}
	return
}

func (rc *RunningContainer) createProxy(insideDocker bool) (*httputil.ReverseProxy, error) {
	if rc.proxy != nil {
		log.Println("Proxy already running")
		return nil, nil
	}
	if len(rc.container.Ports) > 0 && len(rc.container.Names) > 0 && strings.HasPrefix(rc.container.Names[0], "/") {
		var hostname string
		if insideDocker {
			hostname = "host.docker.internal"
		} else {
			hostname = "localhost"
		}
		containerUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", hostname, rc.container.Ports[0].PublicPort))
		if err != nil {
			return nil, err
		}
		return httputil.NewSingleHostReverseProxy(containerUrl), nil
	} else {
		// TODO: returns validation error
	}
	return nil, nil
}
