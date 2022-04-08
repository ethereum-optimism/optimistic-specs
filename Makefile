build: submodules opnode contracts
.PHONY: build

submodules:
	# CI will checkout submodules on its own (and fails on these commands)
	if [ -z "$$GITHUB_ENV" ]; then \
		git submodule init; \
		git submodule update; \
	fi
.PHONY: submodules

opnode:
	go build -o ./bin/op ./opnode/cmd
.PHONY: opnode

contracts:
	cd ./packages/contracts && yarn install && yarn build
.PHONY: contracts

test:
	cd ./opnode && make test
	cd ./packages/contracts && yarn test
.PHONY: test

clean:
	rm -rf ./bin
.PHONY: clean

devnet-clean: devnet-down
	cd ./ops && docker-compose rm
	docker volume rm ops_l1_data
	docker volume rm ops_l2_data
.PHONY: devnet-clean

devnet-up:
	@test -f ./packages/contracts/artifacts/contracts/L1/DepositFeed.sol/DepositFeed.json
	@test -f ./packages/contracts/artifacts/contracts/L2/L1Block.sol/L1Block.json
	@test -f ./packages/contracts/artifacts/contracts/L2/Withdrawer.sol/Withdrawer.json
	@(cd ./ops && \
		DEPOSIT_FEED_BYTECODE=$(shell cat ./packages/contracts/artifacts/contracts/L1/DepositFeed.sol/DepositFeed.json | jq .deployedBytecode) \
			L1_BLOCK_INFO_BYTECODE=$(shell cat ./packages/contracts/artifacts/contracts/L2/L1Block.sol/L1Block.json | jq .deployedBytecode) \
			WITHDRAWER_BYTECODE=$(shell cat ./packages/contracts/artifacts/contracts/L2/Withdrawer.sol/Withdrawer.json | jq .deployedBytecode) \
			GENESIS_TIMESTAMP=$(shell date +%s) \
            BUILDKIT_PROGRESS=plain docker-compose up --build)
.PHONY: devnet-up

devnet-down:
	@(cd ./ops && docker-compose down)
.PHONY: devnet-stop
