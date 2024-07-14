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

type Karja struct {
	dockerClient *client.Client
	containers   []RunningContainer
	hostname     string
	// Whether this process is running in Docker
	insideDocker bool
	// The container which is running karja itself (nullable)
	me *RunningContainer
}

func main() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: Use `docker run --cidfile` to detect whether karja is running inside of docker or not
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	var insideDocker bool
	if _, err := os.Stat("/.dockerenv"); err == nil {
		insideDocker = true
	}
	if insideDocker {
		log.Println("Running inside of Docker")
	} else {
		log.Println("Running outside of Docker")
	}

	karja := Karja{dockerClient, []RunningContainer{}, hostname, insideDocker, nil}
	_, err = karja.fetchContainers()
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP server
	assetsFS, err := fs.Sub(assets, "web/dist")
	if err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", karja.handleReverseProxy(http.FileServer(http.FS(assetsFS))))
	mux.Handle("/api/containers", karja.handleReverseProxy(http.HandlerFunc(karja.resolveContainers)))

	go karja.watchContainers(context.TODO())
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (k *Karja) watchContainers(ctx context.Context) {
	for {
		ticker := time.NewTicker(3 * time.Second)
		select {
		case <-ticker.C:
			if err := k.updateContainers(); err != nil {
				log.Println("Failed to fetch containers", err)
			}
			log.Printf("Fetched %d containers", len(k.containers))
		case <-ctx.Done():
			log.Print("Interrupts containers watching")
			break
		}
	}
}

// Forward a request to a reverse proxy based on subdomain
func (k *Karja) handleReverseProxy(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ctr := range k.containers {
			if strings.HasPrefix(r.Host, ctr.Name+".") {
				if ctr.proxy != nil {
					ctr.proxy.ServeHTTP(w, r)
				} else {
					log.Println("Reverse proxy is not set yet")
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

// `GET /api/containers`
func (k *Karja) resolveContainers(w http.ResponseWriter, r *http.Request) {
	var containers []ApiContainer
	for _, ctr := range k.containers {
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
