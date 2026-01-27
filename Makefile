# --- HYDRA-WORM UNIFIED TACTICAL INTERFACE ---
SHELL := /bin/bash
PROJECT_ROOT := $(shell pwd)
COMPOSE_FILE := lab/docker-compose.yml

# --- PATHS ---
C2_DIR       := hydra-c2
C2_SOURCE    := ./main.go
C2_BINARY    := ../lab/hydra-c2
AGENT_BIN    := hydra-agent/target/x86_64-unknown-linux-musl/debug/hydra-agent

# --- TARGETS ---
C2_TARGET    := hydra-c2-lab
ALPHA_TARGET := hydra-agent-alpha
BRAVO_TARGET := hydra-agent-bravo

.PHONY: up down restart patch patch-agent patch-c2 attach-c2 attach-alpha shell block-alpha unblock-alpha help

# --- Lab Management ---
up:
	@echo "[*] Deploying 4-Machine Hydra Lab..."
	docker compose -f $(COMPOSE_FILE) up -d

down:
	@echo "[!] Tearing down Lab Infrastructure..."
	docker compose -f $(COMPOSE_FILE) down

restart:
	docker compose -f $(COMPOSE_FILE) restart

# --- Component Patching ---
patch: patch-c2 patch-agent

patch-agent:
	@echo "[*] Rebuilding Rust Agent (musl target)..."
	cd hydra-agent && cargo build --target x86_64-unknown-linux-musl
	@echo "[*] Hot-swapping binaries in Alpha and Bravo..."
	docker cp $(AGENT_BIN) $(ALPHA_TARGET):/app/hydra-agent
	docker cp $(AGENT_BIN) $(BRAVO_TARGET):/app/hydra-agent
	docker restart $(ALPHA_TARGET) $(BRAVO_TARGET)

patch-c2:
	@echo "[*] Rebuilding Go Orchestrator..."
	cd $(C2_DIR) && CGO_ENABLED=0 GOOS=linux go build -o $(C2_BINARY) $(C2_SOURCE)
	docker cp lab/hydra-c2 $(C2_TARGET):/app/hydra-c2
	docker restart $(C2_TARGET)

# --- Traffic Jam: Tier Blocking ---
# Usage: make block-alpha TIER=4
block-alpha:
	@$(MAKE) toggle-traffic TARGET=$(ALPHA_TARGET) ACTION=DROP

unblock-alpha:
	@$(MAKE) toggle-traffic TARGET=$(ALPHA_TARGET) ACTION=ACCEPT

toggle-traffic:
ifeq ($(TIER),1) # Tier 1: Cloud/HTTP
	docker exec -u 0 $(TARGET) iptables -A OUTPUT -p tcp --dport 8080 -j $(ACTION)
endif
ifeq ($(TIER),4) # Tier 4: ICMP
	docker exec -u 0 $(TARGET) iptables -A OUTPUT -p icmp -j $(ACTION)
endif
ifeq ($(TIER),5) # Tier 5: NTP
	docker exec -u 0 $(TARGET) iptables -A OUTPUT -p udp --dport 123 -j $(ACTION)
endif
ifeq ($(TIER),6) # Tier 6: DNS
	docker exec -u 0 $(TARGET) iptables -A OUTPUT -p udp --dport 53 -j $(ACTION)
endif

# --- Attachment & Logs ---
shell:
	@echo "[*] Attaching to C2 Console... (Ctrl-E then e to detach)"
	docker attach --detach-keys="ctrl-e,e" $(C2_TARGET)

attach-alpha:
	docker logs -f $(ALPHA_TARGET)

help:
	@echo "Hydra Management Commands:"
	@echo "  make up / down      - Manage lab lifecycle"
	@echo "  make patch          - Rebuild and hot-swap all binaries"
	@echo "  make block-alpha    - Block traffic (requires TIER=x)"
	@echo "  make shell          - Attach to Orchestrator TUI"