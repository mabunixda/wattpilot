all: clean fmt test wattpilot_shell wattpilot_exporter

preprocess: fmt
	go generate ./...
	go fmt ./...

fmt:
	go fmt ./...

wattpilot_exporter:
	make -C prometheus all

wattpilot_shell:
	make -C shell all

clean:
	make -C prometheus clean
	make -C shell clean

docker-context:
	(docker buildx ls | grep ^wattpilot_exporter > /dev/null ) && echo "buildx context exists" || docker buildx create --name wattpilot_exporter
	docker buildx use wattpilot_exporter
	docker buildx inspect --bootstrap

docker: docker-prometheus docker-shell

docker-prometheus:
	docker buildx build --platform linux/amd64,linux/arm64 -t mabunixda/wattpilot_exporter --push --build-arg BINARY=prometheus .

docker-shell:
	docker buildx build --platform linux/amd64,linux/arm64 -t mabunixda/wattpilot_shell --build-arg BINARY=shell .

test:
	go test -v ./
