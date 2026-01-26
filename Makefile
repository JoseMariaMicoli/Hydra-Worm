# --- HYDRA-WORM UNIFIED TACTICAL INTERFACE ---
SHELL := /bin/bash
PROJECT_ROOT := $(shell pwd)
AGENT_BIN := hydra-agent/target/x86_64-unknown-linux-musl/debug/hydra-agent
COMPOSE_FILE := lab/docker-compose.yml

# [CORRECTION] Target the Service Name defined in docker-compose.yml
C2_SERVICE := orchestrator
AGENT_CONTAINER := hydra-agent-alpha

.PHONY: up down restart patch-agent patch-c2 logs-c2 logs-agent test-agent shell clean reset help

up:
	export DOCKER_BUILDKIT=1 && docker compose -f $(COMPOSE_FILE) up --build -d
	@$(MAKE) help

patch-agent:
	@echo "[*] Rebuilding Rust Agent (musl target)..."
	cd hydra-agent && cargo build --target x86_64-unknown-linux-musl
	@echo "[*] Injecting binary into $(AGENT_CONTAINER)..."
	docker cp $(AGENT_BIN) $(AGENT_CONTAINER):/app/hydra-agent
	docker restart $(AGENT_CONTAINER)
	@echo "[+] Ghost patched and re-engaged."

# [CORRECTION] Force rebuild of the C2 binary inside the container context
patch-c2:
	@echo "[*] Patching Hydra Orchestrator (Service: $(C2_SERVICE))..."
	docker compose -f $(COMPOSE_FILE) stop $(C2_SERVICE)
	docker compose -f $(COMPOSE_FILE) build --no-cache $(C2_SERVICE)
	docker compose -f $(COMPOSE_FILE) up -d $(C2_SERVICE)
	@echo "[+] Orchestrator redeployed. Attaching to logs..."
	docker compose -f $(COMPOSE_FILE) logs -f $(C2_SERVICE)

shell:
	docker attach hydra-c2-lab

logs-c2:
	docker compose -f $(COMPOSE_FILE) logs -f $(C2_SERVICE)

logs-agent:
	docker logs -f $(AGENT_CONTAINER)

reset: down
	docker system prune -f
	$(MAKE) up

help:
	@echo "--- HYDRA OPERATIONAL COMMANDS ---"
	@echo "make patch-agent  - Hot-swap Rust binary"
	@echo "make patch-c2     - Rebuild & Redeploy C2 [Fixes 404/NTP]"
	@echo "make logs-c2      - Monitor C2 Output"
	@echo "make logs-agent   - Monitor Agent Telemetry"