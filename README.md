Karja
====
[![Docker Image Size (latest tag)](https://img.shields.io/docker/image-size/mtgto/karja/latest)](https://hub.docker.com/r/mtgto/karja/)
[![docker build status](https://github.com/mtgto/karja/actions/workflows/build.yml/badge.svg)](https://github.com/mtgto/karja/actions/workflows/build.yml)

Karja is a HTTP reverse proxy to Docker containers for local web development.

## Screenshot

![Dashboard](assets/dashboard.jpg)

## Features

- **Reverse Proxy:** Act as a HTTP reverse proxy to containers in your Docker
- **Dashboard:** Display running Docker containers
- **One binary:** Work with just one binary file

## Tech Stack

- [Golang](https://go.dev/)
- [Docker Engine Go SDK](https://docs.docker.com/engine/api/sdk/)
- [Svelte](https://svelte.dev/)
- [Pico CSS](https://picocss.com/)

## Usage

There are two ways to run Karja: as a Docker container or outside of Docker.

### 1. As a Docker container

Don't forget to mount `docker.sock` to call Docker API from Karja.

```bash
docker run -d --name karja -v /var/run/docker.sock:/var/run/docker.sock -p 80:9000 mtgto/karja
```

Open http://localhost

### 2. Outside of Docker

```bash
go build
PORT=80 ./karja
```

Open http://localhost

## Try it out

See [examples/docker-compose.yml](examples/docker-compose.yml).

## Environment Variables

Name | Value
---- | -----
PORT | TCP Port number to listen. Default is `9000`.

## Requirements

You can run Karja inside or outside of Docker.
Karja automatically finds the way to your container depends on whether it is inside or outside Docker.

- Inside of Docker (as a Docker container)
  - If Karja and the target container share the same Docker network, determine the port in the following order:
    1. `VIRTUAL_PORT` environment variable in target container
    2. Exported TCP port
    3. Port 80
  - Otherwise, select exported TCP port if exists
- Outside of Docker
  - Select exported TCP port if exists

## Development

Karja consists of a frontend with Svelte and a backend with Golang.
Before you develop backend, you should build web assets.
`go build` embed web assets automatically.

```bash
cd web
pnpm install
pnpm build

cd ..
go build
./karja
```

## Related projects

- [yaichi](https://github.com/mtsmfm/yaichi) Great software that was the basis ideas for Karja such as the dashboard and reverse proxy for running containers in local development.
- [nginx-proxy](https://github.com/nginx-proxy/nginx-proxy) Famous reverse proxy software for docker containers in production. We refer to how nginx-proxy determines the port to the container.

## License

Apache 2.0
