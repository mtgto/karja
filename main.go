package main

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"github.com/docker/docker/client"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed web/dist/*
var assets embed.FS

type ReverseProxyService struct {
	containers []RunningContainer
	// The container which is running karja itself (nullable)
	me *RunningContainer
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
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	var me *RunningContainer
	for _, c := range containers {
		if strings.HasPrefix(c.container.ID, hostname) {
			log.Println(c.container.ID)
			me = &c
			break
		}
	}
	if me == nil {
		log.Println("Karja is running outside of Docker")
		log.Println(hostname)
	} else {
		log.Printf("Karja is running inside of Docker (%s)", me.container.ID)
	}
	assetsFS, err := fs.Sub(assets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}
	service := &ReverseProxyService{containers, me}
	mux := http.NewServeMux()
	mux.Handle("/", service.handleReverseProxy(http.FileServer(http.FS(assetsFS))))
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
