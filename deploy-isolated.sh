#!/bin/sh
# Run on the target host from this package's own directory.
set -eu

cd "$(dirname "$0")"
[ -f .env ] || { echo 'Missing private .env' >&2; exit 1; }

project=$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env | tail -n 1)
port=$(sed -n 's/^TRIAL_HTTP_PORT=//p' .env | tail -n 1)
[ "$project" = 'hris-client-trial' ] || { echo 'Unexpected Compose project name' >&2; exit 1; }
[ "$port" = '13001' ] || { echo 'Unexpected HTTP port' >&2; exit 1; }
[ "$(pwd -P)" = '/apps/hris-client-trial' ] || { echo 'Run only from /apps/hris-client-trial' >&2; exit 1; }

if docker ps -aq --filter "label=com.docker.compose.project=$project" | grep -q .; then
  echo 'A stack with this project name already exists; refusing to change it' >&2
  exit 1
fi
if ss -ltnH "( sport = :$port )" | grep -q .; then
  echo "Port $port is occupied; refusing to bind it" >&2
  exit 1
fi

umask 077
docker compose -f compose.yaml config --quiet
docker compose -f compose.yaml build
docker compose -f compose.yaml up -d postgres redis
docker compose -f compose.yaml run --rm --no-deps --entrypoint ./seeder backend
docker compose -f compose.yaml up -d backend frontend
docker compose -f compose.yaml ps
curl --fail --silent --show-error "http://127.0.0.1:$port/api/v1/health"
echo
echo "Trial stack is listening on 127.0.0.1:$port"
