FROM gcr.io/distroless/static-debian12

COPY karja.linux-arm64 /usr/bin/karja
ENTRYPOINT ["/usr/bin/karja"]
