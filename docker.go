// SPDX-FileCopyrightText: 2024 mtgto <hogerappa@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"log"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
)

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

// Update running containers
func (k *Karja) updateContainers() error {
	containers, err := k.fetchContainers()
	if err != nil {
		return err
	}
	if k.insideDocker && k.me == nil {
		k.findMe(containers)
	}
	for i, rc := range containers {
		index := slices.IndexFunc(k.containers, func(krc RunningContainer) bool {
			return krc.container.ID == rc.container.ID
		})
		if rc.healthy && rc.proxy == nil {
			if index >= 0 && k.containers[index].proxy != nil {
				containers[i].proxy = k.containers[index].proxy
			} else {
				proxy, err := rc.createProxy(k.insideDocker)
				if err != nil {
					log.Fatal("Failed to create proxy:", err)
				}
				containers[i].proxy = proxy
			}
		}
	}
	k.containers = containers
	return nil
}

// Fetch running containers using Docker API
func (k *Karja) fetchContainers() (ret []RunningContainer, err error) {
	containers, err := k.dockerClient.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ctr := range containers {
		// Exclude PublicPort == 0 containers (= not exported)
		if len(ctr.Ports) > 0 && ctr.Ports[0].PublicPort > 0 && ctr.Ports[0].Type == "tcp" && len(ctr.Names) > 0 && strings.HasPrefix(ctr.Names[0], "/") {
			// ctr.Names starts with "/"
			name := strings.TrimPrefix(ctr.Names[0], "/")
			healthy := ctr.State == "running"
			ret = append(ret, RunningContainer{name, healthy, ctr, false, nil})
		}
	}
	return
}

func (k *Karja) findMe(containers []RunningContainer) {
	for _, rc := range containers {
		if strings.HasPrefix(rc.container.ID, k.hostname) {
			log.Printf("Detect the container running karja itself: (%s).", rc.container.ID)
			k.me = &rc
			return
		}
	}
}

func (rc *RunningContainer) createProxy(insideDocker bool) (*httputil.ReverseProxy, error) {
	if rc.proxy != nil {
		log.Println("Proxy already running")
		return nil, nil
	}
	port := rc.container.Ports[0].PublicPort
	if len(rc.container.Ports) > 0 && len(rc.container.Names) > 0 && strings.HasPrefix(rc.container.Names[0], "/") {
		var hostname string
		if insideDocker {
			hostname = "host.docker.internal"
		} else {
			hostname = "localhost"
		}
		containerUrl, err := url.Parse(fmt.Sprintf("http://%s:%d", hostname, port))
		if err != nil {
			return nil, err
		}
		return httputil.NewSingleHostReverseProxy(containerUrl), nil
	} else {
		// TODO: returns validation error
	}
	return nil, nil
}
