FROM ubuntu:jammy

COPY karja.linux-arm64 /usr/bin/karja
ENTRYPOINT ["/usr/bin/karja"]
