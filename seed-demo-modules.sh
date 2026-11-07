#!/bin/sh
set -eu
cd "$(dirname "$0")"
[ -f .env ] || { echo 'Missing .env' >&2; exit 1; }
[ "$(sed -n 's/^COMPOSE_PROJECT_NAME=//p' .env | tail -n 1)" = 'hris-client-trial' ] || { echo 'Unexpected Compose project' >&2; exit 1; }
[ "$(sed -n 's/^TRIAL_HTTP_PORT=//p' .env | tail -n 1)" = '13001' ] || { echo 'Unexpected trial port' >&2; exit 1; }
[ "$(pwd -P)" = '/apps/super-app' ] || { echo 'Run only from /apps/super-app' >&2; exit 1; }
docker compose -f compose.yaml exec -T postgres psql -X -v ON_ERROR_STOP=1 -U trial -d hris_trial < seed-demo-modules.sql
