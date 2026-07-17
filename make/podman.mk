podman/up:
	podman-compose up -d --build --force-recreate --remove-orphans

podman/resource:
	podman-compose up -d redis
	podman ps

podman/down:
	podman-compose down

podman/prune:
	podman system prune -a --volumes
