CONTEXT_DIR := $(shell pwd)

.PHONY: prepare-backend
prepare-backend:
	@echo "Preparing backend environment..."
	cd ${CONTEXT_DIR}/backend && \
	cp sample.env .env  && \
	go mod download 

.PHONY: prepare-frontend
prepare-frontend:
	@echo "Preparing frontend environment..."
	cd ${CONTEXT_DIR}/frontend && \
	cp sample.env .env && \
	bun install && \
	echo "Frontend environment prepared successfully"

.PHONY: start-demo
start-demo:
	@echo "Starting demo environment..."
	@echo "Starting demo services..."
	cd ${CONTEXT_DIR} && docker compose up -d --build
	@echo "Demo environment started successfully"

.PHONY: stop-demo
stop-demo:
	@echo "Stopping demo environment..."
	cd ${CONTEXT_DIR} && docker compose down
	@echo "Demo environment stopped successfully"
