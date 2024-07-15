FROM node:20 AS asset
WORKDIR /node
COPY web/package.json web/pnpm-lock.yaml /node/
RUN npm install -g pnpm
RUN pnpm install
COPY web .
RUN pnpm build

FROM golang:1.22 AS app
WORKDIR /go/src/github.com/mtgto/karja
COPY Makefile go.mod go.sum ./
RUN go mod download
COPY *.go .
COPY --from=asset /node/dist ./web/dist
RUN CGO_ENABLED=0 make

FROM gcr.io/distroless/static-debian12

COPY --from=app /go/src/github.com/mtgto/karja/karja /usr/bin/karja
ENTRYPOINT ["/usr/bin/karja"]
