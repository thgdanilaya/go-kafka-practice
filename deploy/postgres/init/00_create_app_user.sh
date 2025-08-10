#!/usr/bin/env bash

set -euo pipefail

: "${APP_DB_USER:=app}"
: "${APP_DB_PASSWORD:=app}"
: "${APP_DB_NAME:=orders}"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
  CREATE ROLE ${APP_DB_USER} WITH LOGIN PASSWORD '${APP_DB_PASSWORD}';
  CREATE DATABASE ${APP_DB_NAME} OWNER ${APP_DB_USER};
EOSQL