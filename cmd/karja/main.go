package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//go:embed index.html
var html string

type RunningContainer struct {
	// container id
	id string
	// container name like "awesome-web-service"
	name string
	// status of container is healthy
	healthy bool
	proxy   *httputil.ReverseProxy
}

type ReverseProxyService struct {
	// TODO: Store information about connected docker containers
	containers []RunningContainer
}

func main() {
	mux := http.NewServeMux()
	containers, err := fetchContainers()
	if err != nil {
		log.Fatal(err)
	}
	service := &ReverseProxyService{containers}
	mux.Handle("/", service)
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *ReverseProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ctr := range s.containers {
		if strings.HasPrefix(r.Host, ctr.name+".") {
			ctr.proxy.ServeHTTP(w, r)
			return
		}
	}

	html, err := template.New("index").Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	if err := html.Execute(w, nil); err != nil {
		log.Fatal(err)
	}
}

func fetchContainers() (ret []RunningContainer, err error) {
	// TODO: Store client in struct
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, ctr := range containers {
		fmt.Printf("%s %v (status: %s)\n", ctr.ID, ctr.Ports, ctr.Status)
		// ctr.Names starts with "/"
		if len(ctr.Ports) > 0 && len(ctr.Names) > 0 && strings.HasPrefix(ctr.Names[0], "/") {
			containerUrl, err := url.Parse(fmt.Sprintf("http://localhost:%d", ctr.Ports[0].PublicPort))
			if err != nil {
				return nil, err
			}
			id := ctr.ID
			name := strings.TrimPrefix(ctr.Names[0], "/")
			healthy := ctr.State == "running"
			proxy := httputil.NewSingleHostReverseProxy(containerUrl)
			ret = append(ret, RunningContainer{id, name, healthy, proxy})
		}
	}
	return
}
