#!/usr/bin/env bash
set -o pipefail

host="${METABASE_HOST:-http://localhost:3000}"
username="${METABASE_USERNAME:-example@example.com}"
password="${METABASE_PASSWORD:-password}"

echo "Fetching the setup token from $host"
setupToken=$(curl -s --fail "${host}/api/session/properties" | jq -er '."setup-token"')

if [ $? -ne 0 ]; then
  echo "Failed to extract setup token"
  exit $?
fi

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

echo "Sending request to setup metabase"
curl -s --fail \
  -X POST \
  -H "Content-Type: application/json" \
  -d "${request}" \
  "${host}/api/setup"
