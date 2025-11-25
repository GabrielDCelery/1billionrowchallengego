build.docker:
	docker build . \
		-f ./docker/Dockerfile.generator \
		-t 1brc/generator:latest

build.go:
	go build -o go-1brc ./main.go

BRC_NUM_OF_ROWS ?= 1000

generate:
	docker run --rm \
		-v $$(pwd):/srv \
		--user $$(id -u):$$(id -g) \
		-e BRC_NUM_OF_ROWS=$(BRC_NUM_OF_ROWS) \
		1brc/generator:latest
