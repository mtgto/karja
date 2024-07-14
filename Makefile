.PHONY: karja

all: karja

karja:
	go build -o $@

karja.linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -trimpath -o $@
