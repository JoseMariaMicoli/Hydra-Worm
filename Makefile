# --- Hydra-Worm Tactical Makefile ---

up:
	export DOCKER_BUILDKIT=1 && docker compose -f lab/docker-compose.yml up --build -d

# Attach to the interactive Go shell (VaporTrace UI)
c2-shell:
	docker attach hydra-c2-lab

# Monitor multi-tier listener logs (DNS, ICMP, NTP, HTTP)
logs-c2:
	docker logs -f hydra-c2-lab

# Monitor agent mutation and heartbeat logs
logs-agent:
	docker logs -f hydra-agent-alpha

# Verify Agent integrity and version
test-agent:
	docker exec -it hydra-agent-alpha ./hydra-agent --help

# Clean lab artifacts
clean:
	docker compose -f lab/docker-compose.yml down