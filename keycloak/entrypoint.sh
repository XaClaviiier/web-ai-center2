#!/bin/bash
set -euo pipefail
umask 077
# Secrets stay out of image layers, CLI arguments and shell tracing.
export KCRAW_DB_PASSWORD="$(</run/secrets/keycloak_db_password)"
if [[ -z "$KCRAW_DB_PASSWORD" ]]; then
  printf '%s\n' 'Keycloak database password file is empty' >&2
  exit 1
fi
cat /opt/keycloak/conf/keycloak.conf /run/operator/keycloak.conf > /tmp/cms-keycloak.conf
exec /opt/keycloak/bin/kc.sh --config-file=/tmp/cms-keycloak.conf "$@"
