# syntax=docker/dockerfile:1
FROM node:20 AS asset
WORKDIR /node
RUN npm install -g pnpm
RUN --mount=type=bind,source=web/package.json,target=package.json \
    --mount=type=bind,source=web/pnpm-lock.yaml,target=pnpm-lock.yaml \
    --mount=type=cache,target=/root/.local/share/pnpm/store,sharing=locked \
    pnpm install
COPY ["./web", "."]
RUN --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm build

FROM golang:1.22 AS app
WORKDIR /go/src/github.com/mtgto/karja
RUN --mount=type=cache,target=/go/pkg/mod/,sharing=locked \
    --mount=type=bind,source=go.sum,target=go.sum \
    --mount=type=bind,source=go.mod,target=go.mod \
    go mod download -x
COPY ["*.go", "go.mod", "go.sum", "Makefile", "./"]
RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=bind,target=./web/dist,from=asset,source=/node/dist \
    CGO_ENABLED=0 LDFLAGS="-w -s" make

FROM gcr.io/distroless/static-debian12

COPY --from=app ["/go/src/github.com/mtgto/karja/karja", "/usr/bin/karja"]
ENTRYPOINT ["/usr/bin/karja"]
