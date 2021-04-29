WORKDIR := /home/likecoin
MONIKER := rickmak
SEED_NODES := 913bd0f4bea4ef512ffba39ab90eae84c1420862@34.82.131.35:26656

.PHONY: build
build:
	go mod download
	go build -o $(WORKDIR)/liked cmd/liked/main.go
	go build -o $(WORKDIR)/likecli cmd/likecli/main.go

.PHONY: init
init:
	mkdir -p $(WORKDIR)/.liked/config
	curl -OL https://gist.githubusercontent.com/nnkken/1d1b9d4aae4acb3d835dd3150f546d44/raw/4d97fd471b4bf3be8c5475efbc0361f4926e65e5/genesis.json
	mv genesis.json $(WORKDIR)/.liked/config
	$(WORKDIR)/liked --home $(WORKDIR)/.liked init --chain-id likecoin-chain-sheungwan $(MONIKER)
	mkdir -p $(WORKDIR)/.likecli
	$(WORKDIR)/likecli --home $(WORKDIR)/.likecli keys add validator

