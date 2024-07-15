.PHONY: karja

LDFLAGS := $(LDFLAGS)

all: karja

karja:
	go build -ldflags="$(LDFLAGS)" -trimpath -o $@

karja.linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o $@
