# Makefile for deploying and interacting with contracts

# Load environment variables from .env file
-include .env
export $(shell sed 's/=.*//' .env)
export PROJECT_DIR := $(shell pwd)

.PHONY: install
install:
	@git submodule update --init --recursive

.PHONY: prepare-env
prepare-env:
	@echo "Prepare environment" && \
	cd $$PROJECT_DIR/observer && \
	cp sample.env .env && \
	cd $$PROJECT_DIR/relayer && \
	cp sample.env .env && \
	cd $$PROJECT_DIR/verifier && \
	cp sample.env .env && \
	cd $$PROJECT_DIR/backend && \
	cp sample.env .env && \
	cd $$PROJECT_DIR/web && \
	cp sample.env .env && \
	cd $$PROJECT_DIR && \
	cp sample.env .env && \
	echo "Environment prepared successfully"

.PHONY: prepare-deps
prepare-deps:
	cd $$PROJECT_DIR/observer && \
	go mod download && \
	cd $$PROJECT_DIR/backend && \
	go mod download && \
	cd $$PROJECT_DIR/web && \
	pnpm install

.PHONY: deploy-source-chain
# Deploy contracts on source chain
deploy-source-chain:
	@forge clean && \
	IS_DEST=false forge script --via-ir script/Deploy.s.sol --rpc-url $(SRC_RPC_URL) --broadcast --slow

.PHONY: deploy-dest-chain
# Deploy contracts on dest chain
deploy-dest-chain:
	@forge clean && \
	IS_DEST=true forge script --via-ir script/Deploy.s.sol --rpc-url $(DEST_RPC_URL) --broadcast --slow

.PHONY: setup-source-chain
setup-source-chain:
	@forge clean && \
	IS_DEST=false forge script script/Setup.s.sol --rpc-url $(SRC_RPC_URL) --broadcast --slow

.PHONY: setup-dest-chain
setup-dest-chain:
	@forge clean && \
	IS_DEST=true forge script script/Setup.s.sol --rpc-url $(DEST_RPC_URL) --broadcast --slow

.PHONY: setup
setup: setup-source-chain setup-dest-chain

.PHONY: start-local-source-chain
# Start local source chain
start-local-source-chain:
	@anvil --balance 300 -b 1 --host 0.0.0.0 --code-size-limit 1024000

.PHONY: start-local-dest-chain
# Start local dest chain
start-local-dest-chain:
	@anvil --balance 300 -b 1 --chain-id 31338 --host 0.0.0.0 --code-size-limit 1024000 -p 8546

# Deploy contracts on local source chain
deploy-local-source-chain:
	@forge clean && \
	IS_DEST=false forge script --via-ir script/Deploy.s.sol --rpc-url $(SRC_RPC_URL) --broadcast --slow

# Deploy contracts on local dest chain
deploy-local-dest-chain:
	@forge clean && \
	IS_DEST=true forge script --via-ir script/Deploy.s.sol --rpc-url $(DEST_RPC_URL) --broadcast --slow

.PHONY: deploy
deploy: deploy-local-source-chain deploy-local-dest-chain

.PHONY: check-source-balance
# Check balance on source chain
check-source-balance:
	@IS_DEST=false forge script script/Operations.s.sol --rpc-url $(SRC_RPC_URL) --broadcast --slow --sig "checkBalance()"

.PHONY: check-dest-balance
# Check balance on dest chain
check-dest-balance:
	@IS_DEST=true forge script script/Operations.s.sol --rpc-url $(DEST_RPC_URL) --broadcast --slow --sig "checkBalance()"

.PHONY: mint-source-token
# Mint token on source chain
mint-source-token:
	@IS_DEST=false forge script script/Operations.s.sol --rpc-url $(SRC_RPC_URL) --broadcast --slow --sig "mintToken()"

.PHONY: mint-dest-token
# Mint token on dest chain
mint-dest-token:
	@IS_DEST=true forge script script/Operations.s.sol --rpc-url $(DEST_RPC_URL) --broadcast --slow --sig "mintToken()"

.PHONY: transfer
# Transfer token between source chain and dest chain
transfer:
	@IS_DEST=false forge script script/Operations.s.sol --rpc-url $(SRC_RPC_URL) --broadcast --slow --sig "transfer()"

.PHONY: deliver-claim
# Deliver claim to the VSL on dest chain
deliver-claim:
	@echo "Deliver claim to the VSL" && \
	cast send --private-key $(DEST_PRIVATE_KEY) $(DEST_VSL_ADDRESS) "deliverClaim(bytes)" $(CLAIM) \
		--rpc-url $(DEST_RPC_URL)

.PHONY: check-source-logs
check-source-logs:
	@cast logs --rpc-url $(SRC_RPC_URL)

.PHONY: check-dest-logs
check-dest-logs:
	@cast logs --rpc-url $(DEST_RPC_URL)

.PHONY: start-ui-dev
start-ui-dev:
	@cd $$PROJECT_DIR/web && \
	pnpm i && pnpm dev

.PHONY: start-ui-prod
start-ui-prod:
	@cd $$PROJECT_DIR/web && \
	pnpm i && pnpm build && pnpm start

.PHONY: start-backend
start-backend:
	@cd $$PROJECT_DIR/backend && \
	go run ./cmd/main.go

.PHONY: start-relayer
start-relayer:
	@cd $$PROJECT_DIR/relayer && \
	(set -a; . ./.env; set +a; cargo run)

.PHONY: build
# Build
build:
	@forge build

.PHONY: start-observer
# Start observer
start-observer:
	@cd $$PROJECT_DIR/observer && \
	go run ./cmd/main.go

.PHONY: start-verifier
start-verifier:
	@cd $$PROJECT_DIR/verifier && \
	go run ./cmd/main.go

.PHONY: get-last-claim-ts
# Verifier interaction
get-last-claim-ts:
	@cast call --rpc-url $(DEST_RPC_URL) $(DEST_VERIFIER_ADDRESS) "lastVerificationTimestamp()" | \
	python3 -c "import datetime, sys; print(datetime.datetime.utcfromtimestamp(int(sys.stdin.read().strip(), 16)).strftime('%Y-%m-%d %H:%M:%S UTC'))"

.PHONY: get-claim-count
get-claim-count:
	@cast call --rpc-url $(DEST_RPC_URL) $(DEST_VERIFIER_ADDRESS) "claimCount()" | \
	python3 -c "import sys; print(int(sys.stdin.read().strip(), 16))"

.PHONY: get-claim-id
get-claim-id:
	@cast call --rpc-url $(DEST_RPC_URL) $(DEST_VERIFIER_ADDRESS) "claimIds(uint256)" $(word 2, $(MAKECMDGOALS))

.PHONY: get-claim-by-id
get-claim-by-id:
	@cast call --rpc-url $(DEST_RPC_URL) $(DEST_VERIFIER_ADDRESS) "claimIdToClaim(bytes32)" $(word 2, $(MAKECMDGOALS))

.PHONY: docker-build
docker-build:
	docker compose --progress plain build --no-cache

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose --remove-orphans down

.PHONY: integration-test
integration-test:
	@./tests/integration-test.sh

.PHONY: e2e-test
e2e-test:
	cd tests/wormhole/e2e && ./run.sh

.PHONY: e2e-test-headless
e2e-test-headless:
	cd tests/wormhole/e2e && ./run-headless.sh