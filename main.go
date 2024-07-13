package main

import (
	"context"
	_ "embed"
	"github.com/docker/docker/client"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed index.html
var html string

type ReverseProxyService struct {
	// TODO: Store information about connected docker containers
	containers []RunningContainer
}

func main() {
	mux := http.NewServeMux()
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	dc := DockerClient{apiClient}
	containers, err := dc.fetchContainers()
	if err != nil {
		log.Fatal(err)
	}
	service := &ReverseProxyService{containers}
	mux.Handle("/", service)

	go service.watchContainers(context.TODO(), &dc)
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *ReverseProxyService) watchContainers(ctx context.Context, dockerClient *DockerClient) {
	for {
		ticker := time.NewTicker(3 * time.Second)
		select {
		case <-ticker.C:
			containers, _ := dockerClient.fetchContainers()
			// TODO: Update only changed
			s.containers = containers
			log.Print("Fetched containers")
		case <-ctx.Done():
			log.Print("Interrupts containers watching")
			break
		}
	}
}

func (s *ReverseProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, ctr := range s.containers {
		if strings.HasPrefix(r.Host, ctr.Name+".") {
			if ctr.proxy != nil {
				ctr.proxy.ServeHTTP(w, r)
			}
			return
		}
	}

	html, err := template.New("index").Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	if err := html.Execute(w, s.containers); err != nil {
		log.Fatal(err)
	}
}
