build-on-host:
	go build -o espotad ./cmd/espotad

build-with-alpine:
	docker build -t esp-ota-server .
	./copy-binary.sh

# blocking call, use <ctrl>-c to stop the daemon
dev-restart: build-on-host
	./espotad -d data -s :8092
