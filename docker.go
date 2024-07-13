package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"net/http/httputil"
	"net/url"
	"strings"
)

type DockerClient struct {
	client *client.Client
}

type RunningContainer struct {
	// container id
	id string
	// container name like "awesome-web-service"
	Name string
	// status of container is healthy
	healthy bool
	proxy   *httputil.ReverseProxy
}

func (c *DockerClient) fetchContainers() (ret []RunningContainer, err error) {
	containers, err := c.client.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ctr := range containers {
		fmt.Printf("%s %v (status: %s)\n", ctr.ID, ctr.Ports, ctr.Status)
		if len(ctr.Ports) > 0 && len(ctr.Names) > 0 && strings.HasPrefix(ctr.Names[0], "/") {
			containerUrl, err := url.Parse(fmt.Sprintf("http://localhost:%d", ctr.Ports[0].PublicPort))
			if err != nil {
				return nil, err
			}
			id := ctr.ID
			// ctr.Names starts with "/"
			name := strings.TrimPrefix(ctr.Names[0], "/")
			healthy := ctr.State == "running"
			proxy := httputil.NewSingleHostReverseProxy(containerUrl)
			ret = append(ret, RunningContainer{id, name, healthy, proxy})
		}
	}
	return
}
