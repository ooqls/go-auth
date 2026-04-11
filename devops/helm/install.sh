#!/usr/bin/env bash
set -euo pipefail

RELEASE_NAME="${1:-go-auth}"
NAMESPACE="${2:-default}"
CHART_DIR="$(dirname "$0")/go-auth"

PG_PASS=$(openssl rand -hex 16)
VALKEY_PASS=$(openssl rand -hex 16)

echo "Installing $RELEASE_NAME into namespace $NAMESPACE"

helm upgrade --install "$RELEASE_NAME" "$CHART_DIR" \
  --namespace "$NAMESPACE" \
  --create-namespace \
  --set postgres.auth.password="$PG_PASS" \
  --set postgresql.auth.password="$PG_PASS" \
  --set valkey.auth.password="$VALKEY_PASS"

echo ""
echo "Passwords (save these):"
echo "  Postgres: $PG_PASS"
echo "  Valkey:   $VALKEY_PASS"
