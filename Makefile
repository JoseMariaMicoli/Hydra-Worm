# Makefile
up:
	export DOCKER_BUILDKIT=1 && docker compose -f lab/docker-compose.yml up --build -d

test:
	docker exec -it hydra-agent-alpha ./hydra-agent --version