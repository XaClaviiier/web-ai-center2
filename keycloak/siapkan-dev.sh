#!/usr/bin/env bash
# Siapkan Keycloak DEV setelah `docker compose -f compose.dev.yml up -d` (keycloak/README.md langkah 3-4), lewat admin API:
#   - salin client secret ai-center-cms ke secrets/cms-client-secret.txt
#   - buat kunci session 32 byte di secrets/session-key.bin (bila belum ada)
#   - buat akun uji uji-direktur (role direktur), uji-pegawai (role pegawai), uji-tanpa-role (tanpa role CMS)
# Password akun acak, ditulis ke secrets/akun-uji.txt (ter-ignore git), tidak dicetak. Hanya untuk dev lokal.
# Pakai: bash keycloak/siapkan-dev.sh   (butuh curl dan openssl; Keycloak di http://localhost:8180)
set -euo pipefail
D=$(cd "$(dirname "$0")" && pwd)
KC=${KEYCLOAK_URL:-http://localhost:8180}
OPU=$(sed -n 's/^bootstrap-admin-username=//p' "$D/secrets/operator.conf" | tr -d '\r')
OPP=$(sed -n 's/^bootstrap-admin-password=//p' "$D/secrets/operator.conf" | tr -d '\r')
TOK=$(curl -sf -d client_id=admin-cli -d grant_type=password --data-urlencode "username=$OPU" --data-urlencode "password=$OPP" \
  "$KC/realms/master/protocol/openid-connect/token" | sed -E 's/.*"access_token":"([^"]+)".*/\1/')
[ -n "$TOK" ] || { echo "gagal masuk sebagai operator (cek secrets/operator.conf dan Keycloak)" >&2; exit 1; }
H="Authorization: Bearer $TOK"
CID=$(curl -sf -H "$H" "$KC/admin/realms/ai-center/clients?clientId=ai-center-cms" | sed -E 's/^\[\{"id":"([^"]+)".*/\1/')
SEC=$(curl -sf -H "$H" "$KC/admin/realms/ai-center/clients/$CID/client-secret" | sed -nE 's/.*"value":"([^"]+)".*/\1/p')
[ -n "$SEC" ] || SEC=$(curl -sf -X POST -H "$H" "$KC/admin/realms/ai-center/clients/$CID/client-secret" | sed -nE 's/.*"value":"([^"]+)".*/\1/p')
printf '%s' "$SEC" > "$D/secrets/cms-client-secret.txt"
echo "client secret: secrets/cms-client-secret.txt"
[ -f "$D/secrets/session-key.bin" ] || openssl rand -out "$D/secrets/session-key.bin" 32
echo "kunci session: secrets/session-key.bin"

touch "$D/secrets/akun-uji.txt"
for pasangan in uji-direktur:direktur uji-pegawai:pegawai uji-tanpa-role:; do
  U=${pasangan%%:*}; role=${pasangan#*:}
  if grep -q "^$U=" "$D/secrets/akun-uji.txt"; then echo "$U: sudah ada, dilewati"; continue; fi
  P=$(openssl rand -hex 10)
  code=$(curl -s -o /dev/null -w '%{http_code}' -X POST -H "$H" -H 'Content-Type: application/json' "$KC/admin/realms/ai-center/users" \
    -d "{\"username\":\"$U\",\"firstName\":\"Uji\",\"lastName\":\"${role:-tanpa role}\",\"email\":\"$U@example.test\",\"emailVerified\":true,\"enabled\":true,\"credentials\":[{\"type\":\"password\",\"value\":\"$P\",\"temporary\":false}]}")
  [ "$code" = 201 ] || { echo "$U: gagal dibuat ($code)" >&2; continue; }
  printf '%s=%s\n' "$U" "$P" >> "$D/secrets/akun-uji.txt"
  if [ -n "$role" ]; then
    UID_=$(curl -sf -H "$H" "$KC/admin/realms/ai-center/users?username=$U&exact=true" | sed -E 's/^\[\{"id":"([^"]+)".*/\1/')
    ROLE=$(curl -sf -H "$H" "$KC/admin/realms/ai-center/clients/$CID/roles/$role")
    curl -sf -o /dev/null -X POST -H "$H" -H 'Content-Type: application/json' \
      "$KC/admin/realms/ai-center/users/$UID_/role-mappings/clients/$CID" -d "[$ROLE]"
  fi
  echo "$U: dibuat${role:+ dengan role $role}"
done
echo "password akun uji: secrets/akun-uji.txt"
