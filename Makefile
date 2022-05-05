COMPOSEFLAGS=-d
ITESTS_L2_HOST=http://localhost:9545

build: build-go contracts integration-tests
.PHONY: build

build-go: submodules op-node op-proposer op-batcher
.PHONY: build-go

build-ts: submodules contracts integration-tests
.PHONY: build-ts

submodules:
	# CI will checkout submodules on its own (and fails on these commands)
	if [ -z "$$GITHUB_ENV" ]; then \
		git submodule init; \
		git submodule update; \
	fi
.PHONY: submodules

op-node:
	make -C ./op-node op-node
.PHONY: op-node

op-batcher:
	make -C ./op-batcher op-batcher
.PHONY: op-batcher

op-proposer:
	make -C ./op-proposer op-proposer
.PHONY: op-proposer

contracts:
	cd ./packages/contracts && yarn install && yarn build
.PHONY: contracts

integration-tests:
	cd ./packages/integration-tests && yarn install && yarn build:contracts
.PHONY: integration-tests

clean:
	rm -rf ./bin
.PHONY: clean

devnet-up:
	@bash ./ops-bedrock/devnet-up.sh
.PHONY: devnet-up

devnet-down:
	@(cd ./ops-bedrock && GENESIS_TIMESTAMP=$(shell date +%s) docker-compose stop)
.PHONY: devnet-down

devnet-clean:
	rm -rf ./packages/contracts/deployments/devnetL1
	rm -rf ./.devnet
	cd ./ops-bedrock && docker-compose down
	docker volume rm ops_l1_data
	docker volume rm ops_l2_data
.PHONY: devnet-clean

test-unit:
	make -C ./op-node test
	make -C ./op-proposer test
	make -C ./op-batcher test
	make -C ./op-e2e test
	cd ./packages/contracts && yarn test
.PHONY: test-unit

test-integration:
	bash ./ops-bedrock/test-integration.sh \
		./packages/contracts/deployments/devnetL1
.PHONY: test-integration

devnet-genesis:
	bash ./ops-bedrock/devnet-genesis.sh
.PHONY: devnet-genesis
