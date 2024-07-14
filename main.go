package main

import (
	"context"
	_ "embed"
	"encoding/json"
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
	mux := http.NewServeMux()
	mux.Handle("/", service.handleReverseProxy(http.HandlerFunc(service.serveAssets)))
	mux.Handle("/api/containers", service.handleReverseProxy(http.HandlerFunc(service.resolveContainers)))

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

func (s *ReverseProxyService) handleReverseProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ctr := range s.containers {
			if strings.HasPrefix(r.Host, ctr.Name+".") {
				if ctr.proxy != nil {
					ctr.proxy.ServeHTTP(w, r)
				}
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *ReverseProxyService) serveAssets(w http.ResponseWriter, r *http.Request) {
	html, err := template.New("index").Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	if err := html.Execute(w, s.containers); err != nil {
		log.Fatal(err)
	}
}

// docker container structure for JSON API
type ApiContainer struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	PublicPort  uint16 `json:"public_port"`
	PrivatePort uint16 `json:"private_port"`
	Status      string `json:"status"`
	Healthy     bool   `json:"healthy"`
}

func (s *ReverseProxyService) resolveContainers(w http.ResponseWriter, r *http.Request) {
	var containers []ApiContainer
	for _, ctr := range s.containers {
		containers = append(containers, ApiContainer{
			ctr.container.ID,
			ctr.Name,
			ctr.container.Ports[0].PublicPort,
			ctr.container.Ports[0].PrivatePort,
			ctr.container.Status,
			ctr.healthy,
		})
	}
	if err := json.NewEncoder(w).Encode(containers); err != nil {
		log.Fatal(err)
	}
}
