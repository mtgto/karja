package main

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/docker/docker/api/types"
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

type ReverseProxyService struct {
	// TODO: Store information about connected docker containers
	proxy      *httputil.ReverseProxy
	containers []types.Container
}

func main() {
	mux := http.NewServeMux()
	otherContainerUrl, err := url.Parse("http://localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(otherContainerUrl)
	containers, err := fetchContainers()
	if err != nil {
		log.Fatal(err)
	}
	service := &ReverseProxyService{proxy, containers}
	mux.Handle("/", service)
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *ReverseProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ctr := range s.containers {
		// ctr.Names starts with "/"
		if strings.HasPrefix(ctr.Names[0], "/") && strings.HasPrefix(r.Host, strings.TrimPrefix(ctr.Names[0], "/")+".") {
			s.proxy.ServeHTTP(w, r)
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

func fetchContainers() (ret []types.Container, err error) {
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
		if len(ctr.Ports) > 0 {
			ret = append(ret, ctr)
		}
	}
	return
}
