# Karja

Karja is a reverse proxy to running containers for local development.

## Screenshot

![Dashboard](assets/dashboard.jpg)

## Features

- **Reverse Proxy** Act as a HTTP reverse proxy to containers running in your Docker
- **Dashboard** Display running Docker containers
- **Single binary** Works with just one binary file written in Golang

## Usage

### As a Docker container

Don't forget to mount `docker.sock` to call Docker API from container.

```bash
  docker run --rm -v /var/run/docker.sock:/var/run/docker.sock -p 80:9000 mtgto/karja
```

Open http://localhost

### Standalone (Docker Outside)

```bash
  go build
  PORT=80 ./karja
```

Open http://localhost

## Related projects

- [yaichi](https://github.com/mtsmfm/yaichi) It has almost the same purpose and uses [ngx_mruby](https://ngx.mruby.org/). It supports containers that don't publish ports

## License

Apache 2.0
