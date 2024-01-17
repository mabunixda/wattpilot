all: fmt wattpilot_shell wattpilot_exporter

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

docker:
	make -C prometheus docker
