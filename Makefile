build:
	docker build . \
		-f ./docker/Dockerfile.generator \
		-t 1brc/generator:latest

BRC_NUM_OF_ROWS ?= 1000

generate:
	docker run --rm \
		-v $$(pwd):/srv \
		--user $$(id -u):$$(id -g) \
		-e BRC_NUM_OF_ROWS=$(BRC_NUM_OF_ROWS) \
		1brc/generator:latest
