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
	hostname   string
	// Whether this process is running in Docker
	insideDocker bool
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
	// TODO: Use `docker run --cidfile` to detect whether karja is running inside of docker or not
	var insideDocker bool
	if _, err := os.Stat("/.dockerenv"); err == nil {
		insideDocker = true
	}
	if insideDocker {
		log.Println("Karja is running inside of Docker.")
	} else {
		log.Println("Karja is running outside of Docker.")
	}
	assetsFS, err := fs.Sub(assets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}
	service := &ReverseProxyService{containers, hostname, insideDocker, nil}
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
			if s.insideDocker && s.me == nil {
				s.findMe(containers)
			}
			for i, container := range s.containers {
				if container.healthy && container.proxy == nil {
					proxy, err := container.createProxy(s.insideDocker)
					if err != nil {
						log.Fatal("Failed to create proxy:", err)
					}
					s.containers[i].proxy = proxy
				}
			}
			log.Print("Fetched containers")
		case <-ctx.Done():
			log.Print("Interrupts containers watching")
			break
		}
	}
}

func (s *ReverseProxyService) findMe(containers []RunningContainer) {
	for _, c := range containers {
		if strings.HasPrefix(c.container.ID, s.hostname) {
			log.Printf("Detect the container running karja itself: (%s).", c.container.ID)
			s.me = &c
			return
		}
	}
}

func (s *ReverseProxyService) handleReverseProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ctr := range s.containers {
			if strings.HasPrefix(r.Host, ctr.Name+".") {
				if ctr.proxy != nil {
					ctr.proxy.ServeHTTP(w, r)
				} else {
					log.Println("Reverse proxy is not set yet.")
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
		// Ignore container which does not export port(s).
		if ctr.container.Ports[0].PublicPort > 0 {
			containers = append(containers, ApiContainer{
				ctr.container.ID,
				ctr.Name,
				ctr.container.Ports[0].PublicPort,
				ctr.container.Ports[0].PrivatePort,
				ctr.container.Status,
				ctr.healthy,
			})
		}
	}
	if err := json.NewEncoder(w).Encode(containers); err != nil {
		log.Fatal(err)
	}
}
