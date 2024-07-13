package main

import "net/http/httputil"

type RunningContainer struct {
	// container id
	id string
	// container name like "awesome-web-service"
	Name string
	// status of container is healthy
	healthy bool
	proxy   *httputil.ReverseProxy
}
