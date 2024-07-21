// SPDX-FileCopyrightText: 2024 mtgto <hogerappa@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"log"
	"net/http/httputil"
	"net/url"
	"slices"
	"strconv"
	"strings"
)

type RunningContainer struct {
	// container name like "awesome-web-service"
	Name string
	// status of container is healthy
	healthy   bool
	container types.Container
	// whether established docker network connection between target and karja.
	// always false when karja is running outside of Docker.
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
				proxy, err := k.createProxy(rc)
				if err != nil {
					log.Printf("Failed to create proxy for %s: %v", rc.Name, err)
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
		// ctr.Names starts with "/"
		if len(ctr.Names) == 0 && !strings.HasPrefix(ctr.Names[0], "/") {
			continue
		}
		// TODO: In insideDocker, exclude container which does not share docker network with karja
		if k.insideDocker {
			name := strings.TrimPrefix(ctr.Names[0], "/")
			healthy := ctr.State == "running"
			ret = append(ret, RunningContainer{name, healthy, ctr, false, nil})
		} else {
			// Exclude PublicPort == 0 containers (= not exported)
			if len(ctr.Ports) > 0 && ctr.Ports[0].PublicPort > 0 && ctr.Ports[0].Type == "tcp" {
				// ctr.Names starts with "/"
				name := strings.TrimPrefix(ctr.Names[0], "/")
				healthy := ctr.State == "running"
				ret = append(ret, RunningContainer{name, healthy, ctr, false, nil})
			}
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

// return remote proxy url contains host and port
func (k *Karja) decideRoute(dest RunningContainer) (*url.URL, error) {
	if !k.insideDocker {
		if len(dest.container.Ports) > 0 && dest.container.Ports[0].PublicPort > 0 && dest.container.Ports[0].Type == "tcp" {
			return url.Parse(fmt.Sprintf("http://localhost:%d", dest.container.Ports[0].PublicPort))
		} else {
			return nil, errors.New("no exported port")
		}
	}
	if k.me == nil {
		log.Println("Cannot decide the route before my container is found")
		return nil, nil
	}
	for _, network1 := range k.me.container.NetworkSettings.Networks {
		for _, network2 := range dest.container.NetworkSettings.Networks {
			if network1.NetworkID == network2.NetworkID {
				info, err := k.dockerClient.ContainerInspect(context.Background(), dest.container.ID)
				if err != nil {
					return nil, err
				}
				// whether dest container has env PORT
				for _, pair := range info.Config.Env {
					kv := strings.Split(pair, "=")
					if len(kv) == 2 && kv[0] == "VIRTUAL_PORT" {
						if port, err := strconv.Atoi(kv[1]); err == nil {
							return url.Parse(fmt.Sprintf("http://%s:%d", network2.IPAddress, port))
						}
					}
				}
				if len(dest.container.Ports) > 0 && dest.container.Ports[0].PublicPort > 0 && dest.container.Ports[0].Type == "tcp" {
					return url.Parse(fmt.Sprintf("http://host.docker.internal:%d", dest.container.Ports[0].PublicPort))
				}
				return url.Parse(fmt.Sprintf("http://%s", network2.IPAddress))
			}
		}
	}
	if len(dest.container.Ports) > 0 && dest.container.Ports[0].PublicPort > 0 && dest.container.Ports[0].Type == "tcp" {
		return url.Parse(fmt.Sprintf("http://host.docker.internal:%d", dest.container.Ports[0].PublicPort))
	} else {
		return nil, errors.New("no route found")
	}
}

func (k *Karja) createProxy(dest RunningContainer) (*httputil.ReverseProxy, error) {
	if dest.proxy != nil {
		log.Println("Proxy already running")
		return nil, nil
	}
	containerUrl, err := k.decideRoute(dest)
	if err != nil {
		return nil, err
	} else if containerUrl != nil {
		return httputil.NewSingleHostReverseProxy(containerUrl), nil
	}
	return nil, nil
}
