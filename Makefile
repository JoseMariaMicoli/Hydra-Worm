# --- HYDRA-WORM UNIFIED TACTICAL INTERFACE ---
SHELL := /bin/bash
PROJECT_ROOT := $(shell pwd)
COMPOSE_FILE := lab/docker-compose.yml

# --- PATHS ---
C2_DIR       := hydra-c2
C2_SOURCE    := ./main.go
C2_BINARY    := ../lab/hydra-c2
AGENT_BIN    := hydra-agent/target/x86_64-unknown-linux-musl/debug/hydra-agent

# --- DYNAMIC RECON ---
# Target service names exactly as defined in docker-compose.yml
C2_NAME      := hydra-c2-lab
AGENT_NAME   := hydra-agent-alpha

# Dynamically find names if running, otherwise use hardcoded defaults
C2_CONTAINER    := $(shell docker ps --filter "label=com.docker.compose.service=orchestrator" --format "{{.Names}}" | head -n 1)
AGENT_CONTAINER := $(shell docker ps --filter "label=com.docker.compose.service=target-alpha" --format "{{.Names}}" | head -n 1)

ifeq ($(C2_CONTAINER),)
    C2_TARGET := $(C2_NAME)
else
    C2_TARGET := $(C2_CONTAINER)
endif

ifeq ($(AGENT_CONTAINER),)
    AGENT_TARGET := $(AGENT_NAME)
else
    AGENT_TARGET := $(AGENT_CONTAINER)
endif

.PHONY: up down restart patch-agent patch-c2 logs-c2 logs-agent shell reset help

up:
	@echo "[*] Deploying Hydra Lab Infrastructure..."
	export DOCKER_BUILDKIT=1 && docker compose -f $(COMPOSE_FILE) up --build -d
	@$(MAKE) help

down:
	@echo "[!] Terminating Lab Infrastructure..."
	docker compose -f $(COMPOSE_FILE) down

patch-agent:
	@echo "[*] Rebuilding Rust Agent (musl target)..."
	cd hydra-agent && cargo build --target x86_64-unknown-linux-musl
	@echo "[*] Injecting binary into $(AGENT_TARGET)..."
	docker cp $(AGENT_BIN) $(AGENT_TARGET):/app/hydra-agent
	docker restart $(AGENT_TARGET)
	@echo "[+] Agent $(AGENT_TARGET) re-engaged."

patch-c2:
	@echo "[*] Syncing Go dependencies and building Orchestrator..."
	cd $(C2_DIR) && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o $(C2_BINARY) $(C2_SOURCE)
	@echo "[*] Injecting binary into $(C2_TARGET)..."
	docker cp lab/hydra-c2 $(C2_TARGET):/app/hydra-c2
	docker restart $(C2_TARGET)
	@echo "[+] Orchestrator $(C2_TARGET) patched successfully."

shell:
	docker attach $(C2_TARGET)

logs-c2:
	docker compose -f $(COMPOSE_FILE) logs -f orchestrator

logs-agent:
	docker logs -f $(AGENT_TARGET)

reset: down
	@echo "[!] Performing Hard Reset..."
	docker system prune -f
	$(MAKE) up

help:
	@echo ""
	@echo "--- HYDRA OPERATIONAL COMMANDS ---"
	@echo "make up           - Start the lab and build images"
	@echo "make patch-agent  - Hot-swap Rust binary"
	@echo "make patch-c2     - Hot-swap Go binary"
	@echo "make shell        - Attach to C2 Console"
	@echo "----------------------------------"