# --- HYDRA-WORM UNIFIED TACTICAL INTERFACE ---
# Integrity Status: CHARLIE_INTEGRATION_SYNC
# Focus: Automated C2 sync and Pristine Target (Charlie) recognition.

SHELL := /bin/bash
PROJECT_ROOT := $(shell pwd)
COMPOSE_FILE := lab/docker-compose.yml

# --- PATHS ---
C2_DIR          := hydra-c2
C2_SOURCE       := ./main.go
C2_BINARY       := ../lab/hydra-c2
C2_DIST_DIR     := $(C2_DIR)/bin
LAB_BIN_DIR     := ../lab/bin
AGENT_BIN_SRC   := hydra-agent/target/x86_64-unknown-linux-musl/debug/hydra-agent

# --- TIER PORTS ---
PORT_CLOUD      := 8080
PORT_HTTPS      := 443
PORT_DNS        := 53
PORT_NTP        := 123

.PHONY: up down reboot restart patch-agent patch-c2 attach-c2 shell stress-test provision-tools block-tier-cloud unblock-all watch-persistence clean-all

# --- Lab Management (Auto-Sync Enabled) ---
up:
	@echo "[*] Deploying Hydra Infrastructure..."
	docker compose -f $(COMPOSE_FILE) up -d --remove-orphans
	@echo "[*] Stabilizing (3s)..."
	@sleep 3
	@$(MAKE) provision-tools
	@echo "[!] Enforcing Version Parity..."
	@$(MAKE) patch-c2
	@$(MAKE) patch-agent

down:
	@echo "[!] Tearing down Lab Infrastructure..."
	docker compose -f $(COMPOSE_FILE) down --remove-orphans

# Hard Reset: Total Wipe and Redeploy with Charlie
reboot:
	@echo "[!] INITIATING HARD RESET: Total Infrastructure Purge..."
	@$(MAKE) down
	@echo "[*] Cold-Booting Lab Environment..."
	@$(MAKE) up
	@echo "[+] Hard Reset Complete. Charlie is deployed as pristine target."

# --- Component Patching & Robust Discovery ---
provision-tools:
	@echo "[*] Installing iptables (Bravo-Only)..."
	@BRAVO_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "target-bravo" | awk '{print $$1}'); \
	if [ ! -z "$$BRAVO_ID" ]; then docker exec $$BRAVO_ID apk add --no-cache iptables; fi

patch-agent:
	@echo "[*] Rebuilding Modular Rust Agent..."
	cd hydra-agent && cargo build --target x86_64-unknown-linux-musl
	@mkdir -p $(C2_DIST_DIR)
	@mkdir -p $(LAB_BIN_DIR)
	@echo "[*] Syncing binary to weapon cache (lab/bin/)..."
	cp $(AGENT_BIN_SRC) $(C2_DIST_DIR)/hydra-agent
	cp $(AGENT_BIN_SRC) $(LAB_BIN_DIR)/hydra-agent
	@echo "[*] Identifying containers..."
	@ALPHA_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "target-alpha" | awk '{print $$1}'); \
	BRAVO_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "target-bravo" | awk '{print $$1}'); \
	if [ -z "$$ALPHA_ID" ] || [ -z "$$BRAVO_ID" ]; then \
		echo "[!] ERROR: Alpha or Bravo container ID not found."; \
		exit 1; \
	fi; \
	echo "[+] Injecting into Alpha ($$ALPHA_ID) and Bravo ($$BRAVO_ID)..."; \
	docker cp $(AGENT_BIN_SRC) $$ALPHA_ID:/app/hydra-agent; \
	docker cp $(AGENT_BIN_SRC) $$BRAVO_ID:/app/hydra-agent; \
	echo "[!] Note: Charlie remains pristine (no agent injection)."; \
	docker restart $$ALPHA_ID $$BRAVO_ID

patch-c2:
	@echo "[*] Rebuilding Go Orchestrator..."
	cd $(C2_DIR) && CGO_ENABLED=0 GOOS=linux go build -o $(C2_BINARY) $(C2_SOURCE)
	@C2_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "orchestrator" | awk '{print $$1}'); \
	if [ ! -z "$$C2_ID" ]; then \
		docker cp lab/hydra-c2 $$C2_ID:/app/hydra-c2; \
		docker restart $$C2_ID; \
	fi

# --- Network Warfare & Monitoring ---
block-tier-cloud:
	@echo "[!] Dropping Tier 1 (HTTP/8080) on Bravo..."
	@BRAVO_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "target-bravo" | awk '{print $$1}'); \
	docker exec $$BRAVO_ID iptables -A OUTPUT -p tcp --dport $(PORT_CLOUD) -j DROP

unblock-all:
	@echo "[*] Restoring network on Bravo..."
	@BRAVO_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "target-bravo" | awk '{print $$1}'); \
	docker exec $$BRAVO_ID iptables -F

clean-all:
	@echo "[!] SCORCHED EARTH: Removing Hydra containers and pruning networks..."
	docker ps -a --filter "name=target" -q | xargs -r docker rm -f
	docker network prune -f

shell:
	@C2_ID=$$(docker ps --format "{{.ID}} {{.Names}}" | grep -i "orchestrator" | awk '{print $$1}'); \
	docker attach --detach-keys="ctrl-e,e" $$C2_ID