CONTEXT_DIR := $(shell pwd)

.PHONY: start
start: 
	@echo "Start all services" && \
	docker compose up -d --build && \
	echo "All services started successfully"

.PHONY: stop 
stop:
	@echo "Stop all services" && \
	docker compose down && \
	echo "All services stopped successfully"



.PHONY: prepare-envs
prepare-envs:
	@echo "Prepare environment" && \
	cd ${CONTEXT_DIR}/claim-submitter && \
	cp sample.env .env && \
	cd ${CONTEXT_DIR}/claim-verifier && \
	cp sample.env .env && \
	echo "Environment prepared successfully"


.PHONY: prepare-integration-test-envs
prepare-integration-test-envs:
	@echo "Prepare integration test environment" && \
	cd ${CONTEXT_DIR}/claim-submitter && \
	cp integration_test.env .env && \
	cd ${CONTEXT_DIR}/claim-verifier && \
	cp integration_test.env .env && \
	echo "Environment prepared successfully"

.PHONY: generate-header-signer-keys
generate-header-signer-keys:
	@bash -c '\
	echo "Generating private key"; \
	openssl ecparam -genkey -name secp256k1 -rand /dev/random -out priv.pem; \
	echo "Generating public key"; \
	openssl ec -in priv.pem -pubout -out pub.pem; \
	echo "Extracting private key (32 bytes)"; \
	openssl ec -in priv.pem -outform DER | tail -c +8 | head -c 32 | xxd -p -c 32 > priv.key; \
	echo "Extracting uncompressed public key (65 bytes)"; \
	openssl ec -in priv.pem -pubout -outform DER | tail -c 65 | xxd -p -c 65 > pub_uncompressed.key; \
	echo "Extracting compressed public key (33 bytes)"; \
	openssl ec -in priv.pem -pubout -outform DER -conv_form compressed | tail -c 33 | xxd -p -c 33 > pub_compressed.key; \
	PRIV_KEY=$$(cat priv.key); \
	PUB_KEY_UNCOMPRESSED=$$(cat pub_uncompressed.key); \
	PUB_KEY_COMPRESSED=$$(cat pub_compressed.key); \
	echo "============= GENERATED KEYS ============="; \
	echo "HEADER_SIGNER_PRIVATE_KEY=0x$$PRIV_KEY"; \
	echo "HEADER_SIGNER_PUBLIC_KEY=0x$$PUB_KEY_COMPRESSED"; \
	echo "==========================================="; \
	echo "Cleaning up temporary files..."; \
	rm -f priv.pem pub.pem priv.key pub_uncompressed.key pub_compressed.key; \
	echo "Keys generated successfully"; \
	'

.PHONY: generate-submitter-verifier-keys
generate-submitter-verifier-keys:
	@echo "============= GENERATED ETH KEYS (CAST) ============="
	@echo "SUBMITTER Keys:"
	@SUBMITTER_PRIVATE_KEY=$$(cast wallet new | grep "Private key:" | cut -d' ' -f3 | sed 's/^0x//'); \
	SUBMITTER_ADDRESS=$$(cast wallet address 0x$$SUBMITTER_PRIVATE_KEY); \
	echo "SUBMITTER_PRIVATE_KEY=$$SUBMITTER_PRIVATE_KEY"; \
	echo "SUBMITTER_ADDRESS=$$SUBMITTER_ADDRESS"; \
	echo ""; \
	echo "VERIFIER Keys:"; \
	VERIFIER_PRIVATE_KEY=$$(cast wallet new | grep "Private key:" | cut -d' ' -f3 | sed 's/^0x//'); \
	VERIFIER_ADDRESS=$$(cast wallet address 0x$$VERIFIER_PRIVATE_KEY); \
	echo "VERIFIER_PRIVATE_KEY=$$VERIFIER_PRIVATE_KEY"; \
	echo "VERIFIER_ADDRESS=$$VERIFIER_ADDRESS"
	@echo "===================================================="


.PHONY: integration-test
integration-test:
	@echo "Preparing and running integration tests" && \
	$(MAKE) prepare-integration-test-envs && \
	docker compose up --build -d && \
	echo "Integration tests running for 30 seconds..." && \
	sleep 60 && \
	echo "Extracting logs before stopping containers..." && \
	./tests/check-test-result.sh && \
	echo "Stopping containers after log extraction..." && \
	docker compose down && \
	echo "Test validation completed successfully"
