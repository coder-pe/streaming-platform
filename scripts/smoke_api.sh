#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080/api}"
PASSWORD="${PASSWORD:-Test1234!}"
NEW_PASSWORD="${NEW_PASSWORD:-Test1234!!}"
EMAIL="smoke_$(date +%s)@example.com"
FIRST_NAME="Smoke"
LAST_NAME="Tester"

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required for this smoke test"
  exit 1
fi

request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local token="${4:-}"

  local tmp_body
  tmp_body="$(mktemp)"

  local auth_header=()
  if [[ -n "$token" ]]; then
    auth_header=(-H "Authorization: Bearer $token")
  fi

  local status
  if [[ -n "$body" ]]; then
    status=$(curl -sS -o "$tmp_body" -w "%{http_code}" -X "$method" \
      "$BASE_URL$path" \
      -H "Content-Type: application/json" \
      "${auth_header[@]}" \
      -d "$body")
  else
    status=$(curl -sS -o "$tmp_body" -w "%{http_code}" -X "$method" \
      "$BASE_URL$path" \
      "${auth_header[@]}")
  fi

  echo "$status"
  cat "$tmp_body"
  rm -f "$tmp_body"
}

print_result() {
  local name="$1"
  local status="$2"
  if [[ "$status" =~ ^2 ]]; then
    echo "[OK] $name ($status)"
  else
    echo "[FAIL] $name ($status)"
  fi
}

echo "== Smoke API test =="
echo "BASE_URL: $BASE_URL"
echo "EMAIL: $EMAIL"

echo
echo "1) Register"
register_payload=$(jq -n --arg email "$EMAIL" --arg password "$PASSWORD" --arg first "$FIRST_NAME" --arg last "$LAST_NAME" \
  '{email:$email,password:$password,firstName:$first,lastName:$last}')
register_out="$(request POST /auth/register "$register_payload")"
register_status="$(echo "$register_out" | head -n1)"
register_body="$(echo "$register_out" | tail -n +2)"
print_result "register" "$register_status"

echo
echo "2) Login"
login_payload=$(jq -n --arg email "$EMAIL" --arg password "$PASSWORD" \
  '{email:$email,password:$password}')
login_out="$(request POST /auth/login "$login_payload")"
login_status="$(echo "$login_out" | head -n1)"
login_body="$(echo "$login_out" | tail -n +2)"
print_result "login" "$login_status"
TOKEN="$(echo "$login_body" | jq -r '.token // empty')"
USER_ID="$(echo "$login_body" | jq -r '.user.id // empty')"
if [[ -z "$TOKEN" || -z "$USER_ID" ]]; then
  echo "ERROR: missing token or user id in login response"
  exit 1
fi

echo
echo "3) Public videos"
videos_out="$(request GET '/videos?limit=1')"
videos_status="$(echo "$videos_out" | head -n1)"
videos_body="$(echo "$videos_out" | tail -n +2)"
print_result "get videos" "$videos_status"
VIDEO_ID="$(echo "$videos_body" | jq -r '.videos[0].id // empty')"

echo
echo "4) Protected profile/settings"
profile_status="$(request GET /user/profile '' "$TOKEN" | head -n1)"
print_result "get profile" "$profile_status"

stats_status="$(request GET /user/stats '' "$TOKEN" | head -n1)"
print_result "get user stats" "$stats_status"

settings_status="$(request GET /user/settings '' "$TOKEN" | head -n1)"
print_result "get settings" "$settings_status"

update_profile_payload=$(jq -n --arg first "Smoke2" --arg last "Tester2" --arg bio "bio smoke" \
  '{firstName:$first,lastName:$last,bio:$bio}')
update_profile_status="$(request PUT /user/profile "$update_profile_payload" "$TOKEN" | head -n1)"
print_result "update profile" "$update_profile_status"

change_password_payload=$(jq -n --arg current "$PASSWORD" --arg next "$NEW_PASSWORD" \
  '{current_password:$current,new_password:$next}')
change_password_status="$(request POST /auth/change-password "$change_password_payload" "$TOKEN" | head -n1)"
print_result "change password" "$change_password_status"

if [[ -n "$VIDEO_ID" ]]; then
  echo
  echo "5) Video-linked flows with video_id=$VIDEO_ID"

  view_status="$(request POST "/videos/$VIDEO_ID/view" '{}' "$TOKEN" | head -n1)"
  print_result "increment view" "$view_status"

  fav_add_status="$(request POST "/favorites/$VIDEO_ID" '{}' "$TOKEN" | head -n1)"
  print_result "add favorite" "$fav_add_status"

  fav_list_status="$(request GET /favorites '' "$TOKEN" | head -n1)"
  print_result "get favorites" "$fav_list_status"

  progress_payload='{"position":120,"quality":"720p"}'
  progress_status="$(request POST "/history/$VIDEO_ID/progress" "$progress_payload" "$TOKEN" | head -n1)"
  print_result "update watch progress" "$progress_status"

  history_status="$(request GET /history '' "$TOKEN" | head -n1)"
  print_result "get history" "$history_status"

  continue_status="$(request GET /continue-watching '' "$TOKEN" | head -n1)"
  print_result "get continue watching" "$continue_status"

  fav_remove_status="$(request DELETE "/favorites/$VIDEO_ID" '' "$TOKEN" | head -n1)"
  print_result "remove favorite" "$fav_remove_status"
else
  echo "No public video available, skipping video-linked smoke checks."
fi

echo
echo "Smoke test finished."
