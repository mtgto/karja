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

type ReverseProxyService struct {
	// TODO: Store information about connected docker containers
	proxy *httputil.ReverseProxy
}

func main() {
	mux := http.NewServeMux()
	otherContainerUrl, err := url.Parse("http://localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(otherContainerUrl)
	service := &ReverseProxyService{proxy}
	mux.Handle("/", service)
	fetchContainers()
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *ReverseProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Host, "test.") {
		s.proxy.ServeHTTP(w, r)
	} else {
		html, err := template.New("index").Parse(html)
		if err != nil {
			log.Fatal(err)
		}
		if err := html.Execute(w, nil); err != nil {
			log.Fatal(err)
		}
	}
}

func fetchContainers() {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		// TODO: Fatal -> Wait for a while
		log.Fatal(err)
	}

	for _, ctr := range containers {
		fmt.Printf("%s %v (status: %s)\n", ctr.ID, ctr.Ports, ctr.Status)
	}
}
