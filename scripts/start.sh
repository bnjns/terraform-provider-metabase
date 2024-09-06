#!/usr/bin/env bash
set -eo pipefail

readonly health_limit=180

docker compose up -d "${@}"

readonly services=($(docker ps --filter "name=mb-" --format "{{ .Names }}"))

for i in $(seq 1 "$health_limit")
do
  echo "Test $i/$health_limit:"
  all_healthy=true

  for service in "${services[@]}"; do
    health=$(docker inspect --format='{{.State.Health.Status}}' "$service")

    echo "  $service = $health"

    if [ "${health}" != "healthy" ]; then
      all_healthy=false
    fi
  done

  if [ "$all_healthy" = true ]; then
    echo "Services (${services[@]}) healthy after $i seconds"
    exit 0
  fi

  sleep 1
done

echo "Services (${services[@]}) did not become healthy after $health_limit seconds"
docker compose logs > container_logs.log
exit 1
