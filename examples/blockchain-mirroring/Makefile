CONTEXT_DIR := $(shell pwd)

.PHONY: prepare-env
prepare-env:
	@echo "Prepare environment" && \
	cd ${CONTEXT_DIR}/backend && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/frontend && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/mirroring-geth/claim-submitter && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/mirroring-geth/claim-verifier && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/mirroring-reth/claim-submitter && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/mirroring-reth/claim-verifier && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/btc-mirroring/claim-submitter && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/btc-mirroring/claim-verifier && \
	cp sample.env .env && \
	echo "Environment prepared successfully"

.PHONY: install
install:
	@echo "Install the dependencies" && \
	cd ${CONTEXT_DIR}/backend && \
	go mod download && \
	cd ${CONTEXT_DIR}/mirroring-geth && \
	go mod download && \
	cd ${CONTEXT_DIR}/frontend && \
	bun install && \
	echo "Dependencies installed successfully"

.PHONY: start
start:
	echo "Start all services" && \
	docker compose up -d --build && \
	echo "All services started successfully"

.PHONY: stop
stop:
	echo "Stop all services" && \
	docker compose down && \
	echo "All services stopped successfully"

.PHONY: start-backend
start-backend:
	docker compose up -d --build backend

.PHONY: start-mirroring-geth
start-mirroring-geth:
	docker compose up -d --build geth-claim-submitter geth-claim-verifier

.PHONY: start-mirroring-reth
start-mirroring-reth:
	docker compose up -d --build reth-claim-submitter reth-claim-verifier

.PHONY: start-mirroring-bitcoin
start-mirroring-bitcoin:
	docker compose up -d --build btc-claim-verifier btc-claim-submitter

.PHONY: start-frontend-dev
start-frontend-dev:
	@cd ${CONTEXT_DIR}/frontend && \
	bun run dev

.PHONY: start-frontend-prod
start-frontend-prod:
	docker compose up -d --build frontend

.PHONY: integration-test
integration-test:
	@cd $(CONTEXT_DIR)/tests && \
		docker compose up --build --exit-code-from tester && \
		echo "Integration tests completed successfully" && \
		echo "Cleaning up..." && \
		docker compose down && \
		echo "Cleanup completed" && \
		echo "All tasks completed successfully"

.PHONY: generate-submitter-verifier-keys
generate-submitter-verifier-keys:
	@echo "============= GENERATED ETH KEYS (CAST) ============="
	@echo "SUBMITTER Keys:"
	@SUBMITTER_PRIVATE_KEY=$$(cast wallet new | grep "Private key:" | cut -d' ' -f3); \
	SUBMITTER_ADDRESS=$$(cast wallet address $$SUBMITTER_PRIVATE_KEY); \
	echo "SUBMITTER_PRIVATE_KEY=$$SUBMITTER_PRIVATE_KEY"; \
	echo "SUBMITTER_ADDRESS=$$SUBMITTER_ADDRESS"; \
	echo ""; \
	echo "VERIFIER Keys:"; \
	VERIFIER_PRIVATE_KEY=$$(cast wallet new | grep "Private key:" | cut -d' ' -f3); \
	VERIFIER_ADDRESS=$$(cast wallet address $$VERIFIER_PRIVATE_KEY); \
	echo "VERIFIER_PRIVATE_KEY=$$VERIFIER_PRIVATE_KEY"; \
	echo "VERIFIER_ADDRESS=$$VERIFIER_ADDRESS"
	@echo "===================================================="