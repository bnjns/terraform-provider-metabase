#!/usr/bin/env bash
set -eo pipefail

host="${METABASE_HOST:-http://localhost:3000}"
username="${METABASE_USERNAME:-example@example.com}"
password="${METABASE_PASSWORD:-password}"

setupToken=$(curl -s --fail "${host}/api/session/properties" | jq -r '."setup-token"')

request=$(jq -n "{
  database: null,
  invite: null,
  prefs: {
    allow_tracking: false,
    site_locale: \"en\",
    site_name: \"Example\"
  },
  user: {
    email: \"${username}\",
    first_name: \"Example\",
    last_name: \"User\",
    password: \"${password}\",
    password_confirm: \"${password}\",
    site_name: \"Example\"
  },
  token: \"${setupToken}\"
}")

curl -s --fail \
  -X POST \
  -H "Content-Type: application/json" \
  -d "${request}" \
  "${host}/api/setup"
