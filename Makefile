COMMIT := $(shell git rev-parse --short=8 HEAD)
WORKDIR := /home/fotan
VERSION := fotan-1
CHAIN_ID := likecoin-chain-public-testnet-2
SEED_NODE := c5e678f14219c1f161cb608aaeda37933d71695d@nnkken.dev:31801
GENESIS_URL := https://gist.githubusercontent.com/nnkken/a4eff0359b1acd816aa536bd664eb7ed/raw/207206a952078184b7dea1f152d4068612ef7bd6/genesis.json

MONIKER := rickmak
IDENTITY := 68F4D3A1F6E253AF
WEBSITE := https://www.oursky.comm

.PHONY: build
build:
	go mod download
	go build \
		-ldflags "\
			-X \"github.com/cosmos/cosmos-sdk/version.Name=likecoin-chain\" \
			-X \"github.com/cosmos/cosmos-sdk/version.AppName=liked\" \
			-X \"github.com/cosmos/cosmos-sdk/version.BuildTags=netgo ledger\" \
			-X \"github.com/cosmos/cosmos-sdk/version.Version=${VERSION}\" \
			-X \"github.com/cosmos/cosmos-sdk/version.Commit=${COMMIT}\" \
		" \
		-tags "netgo ledger" \
		-o $(WORKDIR)/liked \
		cmd/liked/main.go

.PHONY: init
init:
	${WORKDIR}/liked --home $(WORKDIR)/.liked init ${MONIKER}
	curl -OL ${GENESIS_URL}
	mv genesis.json $(WORKDIR)/.liked/config
	${WORKDIR}/liked --home $(WORKDIR)/.liked keys add validator

.PHONY: systemd-installation
systemd-installation:
	sudo cp liked.service /etc/systemd/system/liked-fotan.service
	sudo systemctl daemon-reload
	sudo systemctl start liked-fotan.service
	sudo systemctl enable liked-fotan.service

.PHONY: create-validator
create-validator:
	$(eval PUBKEY := $(shell ${WORKDIR}/liked --home $(WORKDIR)/.liked tendermint show-validator))
	${WORKDIR}/liked --home $(WORKDIR)/.liked \
		tx staking create-validator \
		--chain-id ${CHAIN_ID} \
		--from validator \
		--node tcp://127.0.0.1:26657 \
		--commission-max-rate 1.0 \
		--commission-max-change-rate 1.0 \
		--min-self-delegation 1 \
		--moniker ${MONIKER} \
		--identity ${IDENTITY} \
		--website ${WEBSITE}

.PHONY: account-status
account-status:
	$(eval ADDRESS := $(shell ${WORKDIR}/liked --home $(WORKDIR)/.liked keys show validator --address))
	${WORKDIR}/liked --home $(WORKDIR)/.liked \
		query account \
		--chain-id ${CHAIN_ID}} \
		--node tcp://127.0.0.1:26657 \
		${ADDRESS}

.PHONY: status
status:
	curl http://127.0.0.1:26657/status
